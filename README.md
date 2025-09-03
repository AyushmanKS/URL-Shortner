# Go URL Shortener

A simple, in-memory URL shortening service written in Go. This project provides a basic REST API to create short URLs and redirect users to the original long URL.

## Features
- **Shorten Any URL**: Creates a fixed-length short ID for any valid URL.
- **HTTP Redirection**: Redirects short URLs to their original destination.
- **RESTful API**: Provides a clean, JSON-based API for creating links.
- **Lightweight**: Built using only the Go standard library with no external dependencies.

## Tech Stack
- **Go (Golang)**
- Go Standard Library (`net/http`, `encoding/json`, `crypto/md5`)

## Prerequisites
Before you begin, ensure you have the following installed on your system:
- **Go**: Version 1.18 or higher.
- **curl**: A command-line tool for making HTTP requests (or a similar API client like Postman).

## Getting Started

Follow these steps to get the server up and running on your local machine.

#### 1. Save the Code
Save the complete Go source code into a file named `main.go` inside your project directory.

#### 2. Run the Server
Open your terminal, navigate to the project directory, and run the following command:

```bash
go run main.go
```

#### Starting URL Shortener...
Starting URL Shortener...
Server starting on port 3000...

The server is now live and ready to accept requests. Keep this terminal window open.
How to Use the API
You can interact with the server using any API client. The following examples use curl.
1. Shorten a New URL
To shorten a URL, you need to send a POST request to the /shorten endpoint with the original URL in a JSON payload.
Step 1: Create a JSON data file
To avoid command-line quoting issues, it's best to place the JSON payload in a file. Create a new file in your project directory named payload.json and add the following content:

{
    "url": "https://www.google.com/search?q=golang+projects"
}

(You can replace the URL with any link you want to shorten.)
Step 2: Send the request
Now, open a new terminal window and run the following curl command. This command reads the data from the payload.json file.

```bash
curl -X POST -H "Content-Type: application/json" --data "@payload.json" http://localhost:3000/shorten
```

You have successfully created a short URL!
2. Use the Short URL
To use the short URL, simply copy the URL from the response and paste it into your web browser's address bar.
Copy the URL: http://localhost:3000/redirect/c4b1a2d3
Paste it into your browser and press Enter.
You will be instantly redirected to the original long URL (https://www.google.com/search?q=golang+projects).

How It Works
This application uses an in-memory map to store the mapping between short IDs and original URLs. An MD5 hash of the original URL is used to generate an 8-character short ID.
