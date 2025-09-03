# Go URL Shortener

A persistent URL shortening microservice built with Go and PostgreSQL. This project provides a REST API to create permanent short URLs and redirect users to their original destination. The application is deployed and live on Render.

## Live Demo

**Live Demo URL**: [https://url-shortner-pmol.onrender.com](https://url-shortner-pmol.onrender.com)

## Features

-   **Persistent Storage**: Uses a PostgreSQL database to permanently store URL mappings.
-   **Shorten Any URL**: Creates a fixed-length 8-character ID for any valid URL.
-   **HTTP Redirection**: Redirects short URLs to their original destination.
-   **RESTful API**: Provides a clean, JSON-based API for creating links.
-   **Ready for Deployment**: Configured to run on cloud platforms like Render using environment variables.

## Tech Stack

-   **Backend**: Go (Golang)
-   **Database**: PostgreSQL
-   **Go Libraries**: `net/http`, `database/sql`, `github.com/jackc/pgx/v5`
-   **Deployment**: Render

## How to Use the Live API

You can interact with the live, deployed service using any API client. The following examples use `curl`.

### 1. Create a Short URL

To shorten a URL, send a `POST` request to the `/shorten` endpoint of the live service.

First, create a `payload.json` file with the following content, replacing the URL with the one you want to shorten.

```json
{
    "url": "https://github.com/AyushmanKS/URL-Shortner"
}

### Next, run the following command in your terminal. It will send the content of payload.json to the live server.

# Note: This command uses the live URL
```bash
curl -X POST -H "Content-Type: application/json" --data "@payload.json" https://url-shortner-pmol.onrender.com/shorten
```

### The server will respond with the new, permanent short URL.
```bash
{"short_url":"https://url-shortner-pmol.onrender.com/r/xxxxxxxx"}
```

### 2. Use the Short URL
Copy the short_url from the response and paste it into any web browser. You will be instantly redirected to the original destination.

API Endpoints
Endpoint	        Method	    Body (JSON)	                    Success Response (201 Created)	            Description
/shorten	        POST	    {"url": "https://example.com"}	{"short_url": "https://.../r/{id}"}	        Creates a new, permanent short URL.
/r/{short_id}	    GET	        (None)	                        (302 Found Redirect)	                    Redirects to the original URL.

### Local Development Setup
## To run this project on your local machine, you will need Go and PostgreSQL installed.

```bash
git clone https://github.com/AyushmanKS/URL-Shortner.git
cd URL-Shortner
```

### 2. Set Up a Local PostgreSQL Database
Make sure you have a PostgreSQL server running. Create a new database for this project.

### 3. Configure Environment Variable
The application is configured using a DATABASE_URL environment variable. You can set it in your terminal before running the app.

```bash
# Example for PowerShell:
$env:DATABASE_URL="postgres://YOUR_USER:YOUR_PASSWORD@localhost:5432/YOUR_DB_NAME"

# Example for Bash (Linux/macOS/Git Bash):
export DATABASE_URL="postgres://YOUR_USER:YOUR_PASSWORD@localhost:5432/YOUR_DB_NAME"
```

### 4. Install Dependencies
Run go mod tidy to download the required libraries (pgx).

```bash
go mod tidy
```

### 5. Run the Server
Now you can start the server locally.

```bash
go run main.go
```

# The server will start on http://localhost:3000.
### Deployment
# This application is deployed on Render. The deployment process involves:
# Pushing the code to a GitHub repository.
# Creating a PostgreSQL instance on Render and obtaining its Internal Connection URL.
# Creating a Web Service on Render connected to the GitHub repository.
# Setting the DATABASE_URL environment variable in the Web Service settings to the Internal Connection URL from the database.
# Render automatically builds and deploys the application on every push to the main branch.