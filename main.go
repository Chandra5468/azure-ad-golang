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
	"github.com/Chandra5468/azure-ad-golang/models/redis"
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
	/*
		Cors like accept headers, accept request types etc.... for all responses here
	*/

	/* Implement these body parser in golang

	   app.use(bodyParser.urlencoded({extended: false}));
	   app.use(bodyParser.json());
	*/

	// Establish Mongodb connection
	client := mango.CreateConnection()
	// mango.CreateConnection()

	// Establish redis connection
	redisClient := redis.CreateConnection()

	// initiate router

	router := http.NewServeMux()
	router.HandleFunc("GET /v1/hi", func(w http.ResponseWriter, r *http.Request) {
		// for i := 0; i < 1000000000; i++ {
		// 	fmt.Println(i)
		// }
		// mango.Test()
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
