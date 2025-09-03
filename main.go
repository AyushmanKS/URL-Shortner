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
	_ "github.com/jackc/pgx/v5/stdlib"
)

// db holds the connection pool to the database.
var db *sql.DB

// ctx is a global context used for all database operations.
var ctx = context.Background()

// initDB connects to the PostgreSQL database and ensures the `urls` table exists.
func initDB() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	var err error
	db, err = sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Unable to open database connection: %v\n", err)
	}

	if err = db.PingContext(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}
	log.Println("Successfully connected to the database.")

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

// generateShortURL creates a unique 8-character hash from a given URL.
func generateShortURL(originalURL string) string {
	hasher := md5.New()
	hasher.Write([]byte(originalURL))
	return hex.EncodeToString(hasher.Sum(nil))[:8]
}

// createURL inserts a new URL record into the database. If the URL already exists,
// it returns the existing short ID without error.
func createURL(originalURL string) (string, error) {
	shortURL := generateShortURL(originalURL)
	query := "INSERT INTO urls (id, original_url) VALUES ($1, $2)"

	_, err := db.ExecContext(ctx, query, shortURL, originalURL)
	if err != nil {
		var pgErr *pgconn.PgError
		// 23505 is the PostgreSQL error code for "unique_violation".
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return shortURL, nil // The URL already exists, which is not an error for us.
		}
		return "", fmt.Errorf("failed to save to database: %w", err)
	}
	return shortURL, nil
}

// getURL retrieves the original URL for a given short ID.
func getURL(id string) (string, error) {
	var originalURL string
	query := "SELECT original_url FROM urls WHERE id = $1"

	err := db.QueryRowContext(ctx, query, id).Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("URL not found for ID: %s", id)
		}
		return "", fmt.Errorf("error retrieving from database: %w", err)
	}
	return originalURL, nil
}

// ShortUrlHandler handles POST requests to create a new short URL.
func ShortUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body: please provide a JSON object with a 'url' key", http.StatusBadRequest)
		return
	}

	shortURL_ID, err := createURL(data.URL)
	if err != nil {
		log.Printf("Error creating URL: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Construct the response URL dynamically based on the request.
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

// redirectURLHandler handles GET requests to redirect short URLs to their original destination.
func redirectURLHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/r/"):]
	originalURL, err := getURL(id)
	if err != nil {
		http.Error(w, "Link not found or has expired", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusFound)
}

// main is the entry point of the application.
func main() {
	initDB()
	defer db.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.HandleFunc("/shorten", ShortUrlHandler)
	http.HandleFunc("/r/", redirectURLHandler)

	log.Println("Starting server on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error starting the server: %v", err)
	}
}
