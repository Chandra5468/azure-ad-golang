package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Chandra5468/azure-ad-golang/models/mango"
	"github.com/joho/godotenv"
)

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

	// initiate cors

	/* Implement these body parser in golang

	   app.use(bodyParser.urlencoded({extended: false}));
	   app.use(bodyParser.json());
	*/

	// Establish Mongodb connection
	client := mango.CreateConnection()

	// initiate router

	router := http.NewServeMux()
	router.HandleFunc("GET /v1/hi", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode("Namaste modi")
	})
	// Setup and Listen server on specific port mentioned in .env file
	// Handling graceful shutdown
	server := http.Server{
		Addr:    os.Getenv("APP_URL"),
		Handler: router,
	}
	// log.Fatal(http.ListenAndServe(os.Getenv("APP_URL"), router))

	slog.Info("server started", "address", os.Getenv("APP_URL"))

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	<-done // this will be blocking until some os signal is received
	slog.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	// Disconnect Mongo before shutdown
	err = client.Disconnect(context.TODO())
	if err != nil {
		slog.Error("failed to disconnect mongo database during server shutdown")
	} else {
		slog.Info("mongodb successfully disconnected")
	}
	err = server.Shutdown(ctx) // We are giving a time of 5 seconds before shutting down. So that any other running processes can be completed

	if err != nil {
		slog.Error("failed to shutdown server", slog.String("error", err.Error()))
	}

	slog.Info("server shutdown successful")
	/*Command to run

	APP_ENV=local go run main.go

	To Build
	APP_ENV=local go build -o myapp
	*/
}
