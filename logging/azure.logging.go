package logging

import (
	"bytes"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

var esClient *elasticsearch.Client

type AzureESOps struct {
	ActionPerformed   string    `json:"actionPerformed"`
	ClientName        string    `json:"clientName"`
	ActionPerformedAt time.Time `json:"actionPerformedAt"`
	UserId            string    `json:"userId"`
	Successfull       bool      `json:"successfull"`
	Error             error     `json:"error"`
}

// We can also do elasticsearch using http calls.
func CreateESClient() {
	// client, err := esClient.CloudID
	var err error
	esClient, err = elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{os.Getenv("ELASTICSEARCH_HOST_ONE")}, // Did not configure username and password
	})

	// esClient.Indices.Create()

	if err != nil {
		log.Fatal("Error creating elasticsearch client", err)
	} else {
		log.Println("ES client created")
	}
	// return esClient
}

func LogIntoAzureIndex(data *AzureESOps) {
	byteDate, err := json.Marshal(data)

	if err != nil {
		slog.Error("error converting struct to byte for elasticsearch insertion ", "error ", err)
	}
	resp, err := esClient.Index("azuread", bytes.NewReader(byteDate))

	defer resp.Body.Close()

	if err != nil {
		slog.Error("error indexing document in elasticsearch", "error", err)
	} else {
		slog.Info("Document indexed successfully in elasticsearch", "message", "successful")
	}
}
