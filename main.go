// sub2api - A subscription converter API service
// Fork of Wei-Shaw/sub2api
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/yourusername/sub2api/handler"
)

const (
	defaultPort    = 8080
	defaultHost    = "0.0.0.0"
	appName        = "sub2api"
	appVersion     = "dev"
)

func main() {
	var (
		host    string
		port    int
		version bool
	)

	flag.StringVar(&host, "host", getEnvOrDefault("HOST", defaultHost), "Host address to listen on")
	flag.IntVar(&port, "port", getEnvOrDefaultInt("PORT", defaultPort), "Port to listen on")
	flag.BoolVar(&version, "version", false, "Print version information and exit")
	flag.Parse()

	if version {
		fmt.Printf("%s version %s\n", appName, appVersion)
		os.Exit(0)
	}

	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/", handler.IndexHandler)
	mux.HandleFunc("/sub", handler.SubHandler)
	mux.HandleFunc("/health", handler.HealthHandler)

	addr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Starting %s %s on %s", appName, appVersion, addr)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// getEnvOrDefault returns the value of the environment variable named by key,
// or fallback if the variable is not set or empty.
func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// getEnvOrDefaultInt returns the integer value of the environment variable
// named by key, or fallback if the variable is not set, empty, or not a valid integer.
func getEnvOrDefaultInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
