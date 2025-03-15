package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Chandra5468/azure-ad-golang/logging"
	"github.com/Chandra5468/azure-ad-golang/models/mango"
	"github.com/Chandra5468/azure-ad-golang/models/redis"
	v1 "github.com/Chandra5468/azure-ad-golang/routes/v1"
	"github.com/joho/godotenv"
)

var allowedOrigins = map[string]bool{
	"http://localhost:7001": true,
}

// NOTE : Most cors implmentations are for browsers. Not for postman or server-server communication.
/*
	Remove this comment later -- For self understanding.....
	server doesn’t enforce CORS restrictions—browsers do. The server’s job is to set the appropriate headers and optionally reject requests (e.g., for security).
	If you don’t add logic to reject unallowed origins or methods, the request will still reach your handler, and the server will respond.
	The browser will then decide whether to let the client read the response based on the CORS headers.

	EX:
	If you set Access-Control-Allow-Origin: http://localhost:7002 but don’t check the origin and call next.ServeHTTP(w, r) anyway, a request from http://evil.com will still be processed by your server.
	The browser will block the response for http://evil.com because the origin doesn’t match, but your server still did the work unless you explicitly stopped it.
*/
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin == "" { // This is for postman or server to server kind of requests
			next.ServeHTTP(w, r)
			return
		}

		// check if origin is allowed
		if !allowedOrigins[origin] {
			slog.Warn("CORS : unallowed origin", "origin", origin)
			http.Error(w, "CORS: Origin not allowed", http.StatusForbidden)
			return
		}
		// Set cors headers
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// w.Header().Set("Access-Control-Allow-Credentials", "true") // If this is set true, then Allow-Origin can never be *

		if r.Method == http.MethodOptions { // for pre-flight requests
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Loading env file based on environment local, staging, or production
	env := os.Getenv("APP_ENV")

	if env == "" { // if env is not specified then loading local env
		env = "local"
	}

	envFile := fmt.Sprintf("envs/.env.%s", env)
	// load env file
	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("error loading %s file: %v", envFile, err.Error())
	}

	// Establish Mongodb connection
	client := mango.CreateConnection()
	// mango.CreateConnection()

	// Establish redis connection
	redisClient := redis.CreateConnection()

	// Establushing ES Client

	logging.CreateESClient()

	// initiate router

	router := http.NewServeMux()
	// Setup and Listen server on specific port mentioned in .env file
	// Handling graceful shutdown

	// Register routes from each file

	v1.AzureRoutes(router)
	// Wrapping all routes with cors middleware
	handler := corsMiddleware(router)
	server := http.Server{
		Addr:    os.Getenv("APP_URL"),
		Handler: handler,
	}
	// log.Fatal(http.ListenAndServe(os.Getenv("APP_URL"), router))

	slog.Info("server started", "address", os.Getenv("APP_URL"))

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := (server.ListenAndServe())
		if err != nil && err != http.ErrServerClosed {
			slog.Error("server failed to start", slog.String("error", err.Error()))
		}
	}()

	<-done // this will be blocking until some os signal is received
	slog.Info("shutting down server")
	close(done)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	// Disconnect Mongo before shutdown
	// err = client.Disconnect(context.TODO())
	err = client.Disconnect(context.TODO())
	if err != nil {
		slog.Error("failed to disconnect mongo database during server shutdown")
	} else {
		slog.Info("mongodb successfully disconnected")
	}
	err = redisClient.Close()
	if err != nil {
		slog.Error("failed to disconnect redis client to server")
	} else {
		slog.Info("redis disconnected successfully")
	}
	err = server.Shutdown(ctx) // We are giving a time of 5 seconds before shutting down. So that any other running processes can be completed
	if err != nil {
		slog.Error("failed to shutdown server", slog.String("error", err.Error()))
	} else {
		slog.Info("server shutdown successful")
	}

	/*Command to run

	APP_ENV=local go run main.go

	To Build
	APP_ENV=local go build -o myapp
	*/
}
