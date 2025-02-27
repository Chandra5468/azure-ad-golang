package mango

import (
	"log"
	"log/slog"
	"os"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateConnection() {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(os.Getenv("MONGO_URL")).SetServerAPIOptions(serverAPI)

	// Creating a new client and connecting to server
	_, err := mongo.Connect(opts)

	if err != nil {
		log.Fatalf("error while connecting mongo %v ", err.Error())
	}

	slog.Info("Connection to mongo database successful")
}
