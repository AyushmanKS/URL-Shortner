package main

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgconn"
)

// Global variable to hold our database connection pool.
var db *sql.DB

// Global context.
var ctx = context.Background()

// initDB initializes the connection to the PostgreSQL database
// and ensures the necessary 'urls' table exists.
func initDB() {
	// Get the database connection URL from an environment variable.
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Provide a default for local development.
		databaseURL = "postgres://postgres:password@localhost:5432/url_shortener_db"
		log.Println("DATABASE_URL not set, defaulting to local PostgreSQL")
	}

	var err error
	// The "pgx" argument tells database/sql to use the pgx driver.
	db, err = sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	// Ping the database to ensure a connection is established.
	if err = db.Ping(); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	log.Println("Successfully connected to the database.")

	// Create the 'urls' table if it doesn't already exist.
	// This makes the application self-initializing.
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS urls (
		id VARCHAR(8) PRIMARY KEY,
		original_url TEXT NOT NULL,
		creation_date TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`

	if _, err = db.ExecContext(ctx, createTableSQL); err != nil {
		log.Fatalf("Unable to create table: %v\n", err)
	}

	log.Println("Table 'urls' is ready.")
}

func generateShortURL(originalURL string) string {
	hasher := md5.New()
	hasher.Write([]byte(originalURL))
	return hex.EncodeToString(hasher.Sum(nil))[:8]
}

// createURL now saves the URL mapping to the PostgreSQL database.
func createURL(originalURL string) (string, error) {
	shortURL := generateShortURL(originalURL)

	query := "INSERT INTO urls (id, original_url) VALUES ($1, $2)"

	_, err := db.ExecContext(ctx, query, shortURL, originalURL)
	if err != nil {
		// Check if the error is a unique key violation (meaning the URL was already shortened).
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 is the code for unique_violation
			log.Printf("URL %s already exists with ID %s", originalURL, shortURL)
			return shortURL, nil // It's not an error, the link already exists.
		}
		// For any other error, return it.
		return "", fmt.Errorf("failed to save to database: %w", err)
	}

	return shortURL, nil
}

// getURL now retrieves the original URL from the PostgreSQL database.
func getURL(id string) (string, error) {
	var originalURL string

	query := "SELECT original_url FROM urls WHERE id = $1"

	err := db.QueryRowContext(ctx, query, id).Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("URL not found")
		}
		return "", fmt.Errorf("error retrieving from database: %w", err)
	}

	return originalURL, nil
}

// --- HTTP Handlers ---

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the Go + PostgreSQL URL Shortener!")
}

func ShortUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	shortURL_ID, err := createURL(data.URL)
	if err != nil {
		http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
		return
	}

	host := r.Host
	scheme := "http"
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	fullShortURL := fmt.Sprintf("%s://%s/r/%s", scheme, host, shortURL_ID)

	response := struct {
		ShortURL string `json:"short_url"`
	}{ShortURL: fullShortURL}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func redirectURLHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/r/"):]
	originalURL, err := getURL(id)
	if err != nil {
		http.Error(w, "Link not found or has expired", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusFound)
}

func main() {
	// Initialize the database connection.
	initDB()
	// Defer closing the database connection until the application exits.
	defer db.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/shorten", ShortUrlHandler)
	http.HandleFunc("/r/", redirectURLHandler)

	log.Println("Starting server on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error starting the server: %v", err)
	}
}
