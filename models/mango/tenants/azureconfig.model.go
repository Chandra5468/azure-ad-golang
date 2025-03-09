package tenants

import (
	"context"
	"os"

	"github.com/Chandra5468/azure-ad-golang/models/mango"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Products struct {
	TenantId string       `bson:"tenantId"`
	Products ProductsList `bson:"products"`
}
type ProductsList struct {
	Ipaas                IpaasProduct                `bson:"banneripass"`
	AzureActiveDirectory AzureActiveDirectoryProduct `bson:"azureActiveDirectory"`
}

type IpaasProduct struct {
	Type string `bson:"type"`
	URL  string `bson:"url"`
}

type AzureActiveDirectoryProduct struct {
	GrantType    string `bson:"grant_type" json:"grant_type"`
	ClientId     string `bson:"client_id" json:"client_id"`
	ClientSecret string `bson:"client_secret" json:"client_secret"`
	UserName     string `bson:"username" json:"username"`
	Password     string `bson:"password" json:"password"`
	Resource     string `bson:"resource" json:"resource"`
	Scope        string `bson:"scope" json:"scope"`
	URL          string `bson:"url" json:"url"`
}

func GetAzureConfigs(tenantId string, ctx context.Context) (*Products, error) {
	var pds Products
	err := mango.MongoClient.Database(os.Getenv("DEFAULT_DB")).Collection("products").FindOne(ctx, bson.M{"tenantId": "BBH_" + tenantId}).Decode(&pds)
	if err != nil {
		return nil, err
	} else {
		return &pds, err
	}
}
