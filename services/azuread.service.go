package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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
	// var buf bytes.Buffer // what does bytes.Buffer do
	// var buf2 bytes.Reader // what does bytes.Reader do
	// bytes.NewBufferString() or string.NewReader() when to use these and what do they do ?
	data := url.Values{}

	data.Set("grant_type", details.GrantType)
	data.Set("client_id", details.ClientId)
	data.Set("client_secret", details.ClientSecret)
	data.Set("scope", details.Scope)
	data.Set("username", details.UserName)
	data.Set("password", details.Password)
	data.Set("resource", details.Resource)

	req, err := http.NewRequest(http.MethodPost, details.URL, strings.NewReader(data.Encode()))
	// fmt.Println(buf.String())
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
	accessToken := fmt.Sprintf("Bearer %s", azureAccessToken)
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
	err = json.NewDecoder(res.Body).Decode(&getAuthApp.Value)
	if err != nil {
		return nil, err
	}

	return &getAuthApp, nil
}
