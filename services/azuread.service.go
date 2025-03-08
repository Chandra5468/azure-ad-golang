package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Chandra5468/azure-ad-golang/models/mango/tenants"
)

/*
		{
			"id": "fac569a7-5eb8-4df7-b9ae-8886ba87f23a",
			"displayName": "CPH1901",
			"deviceTag": "SoftwareTokenActivated",
			"phoneAppVersion": "6.2501.0191",
			"createdDateTime": null
	    }
*/
type GetMicrosoftAuthResp struct {
	Value MicrosoftAuthInfo `json:"value"`
}
type MicrosoftAuthInfo struct {
	DisplayName     string `json:"displayName"`
	DeviceTag       string `json:"deviceTag"`
	PhoneAppVersion string `json:"phoneAppVersion"`
	// createdDateTime
}

type AccessTokenFromPG struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    string `json:"expires_in"`
	ExtExpiresIn string `json:"ext_expires_in"`
	AccessToken  string `json:"access_token"`
}

func GetAccessTokenPGgrant(details *tenants.AzureActiveDirectoryProduct) (*AccessTokenFromPG, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(&details) // check if it is giving data in json format to NewRequest or not.
	// jsonBody := []byte(`{client_id:}`)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, details.URL, &buf)

	// what does bytes.Buffer do
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var pgAT AccessTokenFromPG
	err = json.NewDecoder(res.Body).Decode(&pgAT)

	if err != nil {
		return nil, err
	}

	return &pgAT, nil
}

func GetMicrosoftAuthenticatorApp(userPrincipalName, azureAccessToken string) (*GetMicrosoftAuthResp, error) {
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/authentication/microsoftAuthenticatorMethods", userPrincipalName)
	var getAuthApp GetMicrosoftAuthResp
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	accessToken := fmt.Sprintf("Bearer %v", azureAccessToken)
	req.Header.Add("Authorization", accessToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	// byteData, err := io.ReadAll(res.Body) // Use it if you want to use unmarshall
	// if err != nil {
	// 	return nil, err
	// }

	// err = json.Unmarshal(byteData, &getAuthApp)
	err = json.NewDecoder(res.Body).Decode(&getAuthApp)
	if err != nil {
		return nil, err
	}

	return &getAuthApp, nil
}
