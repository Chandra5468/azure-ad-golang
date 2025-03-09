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

type LabelConfigs struct {
	Labelling Labelling `json:"labelling" bson:"labelling"`
	Tag       string    `json:"tag" bson:"tag"`
}
type Labelling struct {
	DisplayName                string `json:"displayName" bson:"displayName"`
	Mail                       string `json:"mail" bson:"mail"`
	PersonalEmail              string `json:"personalEmail" bson:"personalEmail"`
	AccountEnabled             string `json:"accountEnabled" bson:"accountEnabled"`
	CreatedDateTime            string `json:"createdDateTime" bson:"createdDateTime"`
	SignInActivity             string `json:"signInActivity" bson:"signInActivity"`
	LastPWDChangeDateTime      string `json:"lastPasswordChangeDateTime" bson:"lastPasswordChangeDateTime"`
	Department                 string `json:"department" bson:"department"`
	MfaMobileNumber            string `json:"mfaMobileNumber" bson:"mfaMobileNumber"`
	MfaAlternativeMobileNumber string `json:"mfaAlternateMobileNumber" bson:"mfaAlternateMobileNumber"`
	MfaOfficeMobileNumber      string `json:"mfaOfficeMobileNumber" bson:"mfaOfficeMobileNumber"`
	MicrosoftAuthenticatorApp  string `json:"microsoftAuthenticatorApp" bson:"microsoftAuthenticatorApp"`
	BusinessPhones             string `json:"businessPhones" bson:"businessPhones"`
	GivenName                  string `json:"givenName" bson:"givenName"`
	JobTitle                   string `json:"jobTitle" bson:"jobTitle"`
	OfficeLocation             string `json:"officeLocation" bson:"officeLocation"`
	PreferredLanguage          string `json:"preferredLanguage" bson:"preferredLanguage"`
	Surname                    string `json:"surname" bson:"surname"`
	UserprincipleName          string `json:"userPrincipalName" bson:"userPrincipalName"`
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

func GetAzureLabelConfigs(tenantId string, ctx context.Context) (*LabelConfigs, error) {
	var lbc LabelConfigs
	err := mango.MongoClient.Database("BBH_"+tenantId).Collection("azureAdLabelConfigs").FindOne(ctx, bson.M{}).Decode(&lbc)
	if err != nil {
		return nil, err
	} else {
		return &lbc, err
	}
}
