package mango

import (
	// "context"

	"log"
	"log/slog"
	"os"

	// "go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

//	type MongoConn struct {
//		conn *mongo.Client
//	}
var MongoClient *mongo.Client

func CreateConnection() *mongo.Client {
	// serverAPI := options.ServerAPI(options.ServerAPIVersion1) // Locally api version to Mongo is not set. So, commenting this
	// opts := options.Client().ApplyURI(os.Getenv("MONGO_URL")).SetServerAPIOptions(serverAPI)
	opts := options.Client().ApplyURI(os.Getenv("MONGO_URL"))

	// Creating a new client and connecting to server
	client, err := mongo.Connect(opts)

	// Should not start the server.
	if err != nil {
		log.Fatalf("error while connecting mongo %v ", err.Error())
	}

	slog.Info("Connection to mongo database successful")

	MongoClient = client
	// return &MongoConn{
	// 	conn: client,
	// }

	// return MongoClient

	return client
}

type Roles struct {
	RoleName       string `bson:"roleName"`
	RoleTitle      string `bson:"roleTitle"`
	LandingPageURL string `bson:"landingPageUrl"`
	TenantId       string `bson:"tenantId"`
	Status         string `bson:"Status"`
}

// func Test() {
// 	var roles Roles
// 	err := MongoClient.Database("Master").Collection("roles").FindOne(context.Background(), bson.M{"roleName": "CIO"}).Decode(&roles)

// 	if err != nil {
// 		log.Fatal("This is find operation error", err)
// 	}
// 	// if cur.Next(context.TODO()) {
// 	// 	cur.Decode(&roles)
// 	// }

// 	fmt.Println("This is data", roles)
// }
