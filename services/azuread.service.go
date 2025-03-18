package services

import (
	"bytes"
	"context"
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

type UserInfo struct {
	DisplayName           string   `json:"displayName"`
	Mail                  string   `json:"mail"`
	AccountEnabled        bool     `json:"accountEnabled"`
	CreatedDateTime       string   `json:"createdDateTime"`
	LastPWDChangeDateTime string   `json:"lastPasswordChangeDateTime"`
	Department            string   `json:"department"`
	BusinessPhones        []string `json:"businessPhones"`
	GivenName             string   `json:"givenName"`
	JobTitle              string   `json:"jobTitle"`
	OfficeLocation        string   `json:"officeLocation"`
	PreferredLanguage     string   `json:"preferredLanguage"`
	Surname               string   `json:"surname"`
	UserprincipleName     string   `json:"userPrincipalName"`
}

type PhoneAuthenticator struct {
	PhoneNumber string `json:"phoneNumber"`
}
type PwdFormat struct {
	NewPassword               string `json:"newPassword"`
	RequireChangeOnNextSignIn bool   `json:"requireChangeOnNextSignIn"`
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

func GetUserInfo(userPrincipalName, azureAccessToken string) (*UserInfo, error) {
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s?$select=displayName,mail,accountEnabled,createdDateTime,lastPasswordChangeDateTime,department,businessPhones,givenName,jobTitle,officeLocation,surname,userPrincipalName", userPrincipalName)

	var uI UserInfo

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

	err = json.NewDecoder(res.Body).Decode(&uI)

	if err != nil {
		return nil, err
	}
	// res.StatusCode ==200
	return &uI, nil
}

func GetPhoneAuthenticatorInfo(userPrincipalName, mobilePhoneId, azureAccessToken string) (string, error) {
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/authentication/phoneMethods/%s", userPrincipalName, mobilePhoneId)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	accessToken := fmt.Sprintf("Bearer %s", azureAccessToken)
	req.Header.Set("Authorization", accessToken)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	var phat PhoneAuthenticator
	defer res.Body.Close()
	// Decode from res.Body to struct
	err = json.NewDecoder(res.Body).Decode(&phat)
	if err != nil {
		return "", err
	}
	return phat.PhoneNumber, nil
}

func MicrosoftAuthDevice(userPrincipalName, azureAccessToken string) (string, error) {
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/authentication/microsoftAuthenticatorMethods", userPrincipalName)
	var getAuthApp GetMicrosoftAuthResp
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	accessToken := fmt.Sprintf("Bearer %s", azureAccessToken)
	req.Header.Add("Authorization", accessToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	// byteData, err := io.ReadAll(res.Body) // Use it if you want to use unmarshall
	// if err != nil {
	// 	return nil, err
	// }

	// err = json.Unmarshal(byteData, &getAuthApp)
	err = json.NewDecoder(res.Body).Decode(&getAuthApp.Value)
	if err != nil {
		return "", err
	}

	return getAuthApp.Value.DisplayName, nil
}

func AzurePwdReset(userPrincipalName, azureAccessToken, newPassword string) (int, error) {
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/authentication/methods/28c10230-6103-485e-b985-444c60001490/resetPassword", userPrincipalName)
	dataType := PwdFormat{
		NewPassword:               newPassword,
		RequireChangeOnNextSignIn: true,
	}
	byteData, err := json.Marshal(&dataType)
	if err != nil {
		return 400, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(byteData))
	if err != nil {
		return 400, err
	}
	req.Header.Add("Content-type", "application/json")
	accessToken := fmt.Sprintf("Bearer %s", azureAccessToken)
	req.Header.Add("Authorization", accessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 400, err
	}
	defer res.Body.Close()
	return res.StatusCode, nil
}

func DeletePhoneAuthenticators(azureAccessToken, userPrincipalName string, ctx context.Context) (uint16, error) {
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/authentication/phoneMethods/3179e48a-750b-4051-897c-87b9720928f7", userPrincipalName)

	req, _ := http.NewRequest(http.MethodDelete, url, nil)

	accessToken := fmt.Sprintf("Bearer %s", azureAccessToken)

	req.Header.Add("Authorization", accessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil && (res.StatusCode == 400 || res.StatusCode == 404) {
		return uint16(res.StatusCode), err
	}
	defer res.Body.Close()

	if res.StatusCode == 204 {
		return uint16(res.StatusCode), nil
	} else {
		return 0, err
	}
}

func DeleteMicrosoftAuthenticators(azureAccessToken, userPrincipalName string, ctx context.Context) (uint16, error) {
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/authentication/microsoftAuthenticatorMethods/3179e48a-750b-4051-897c-87b9720928f7", userPrincipalName)

	req, _ := http.NewRequest(http.MethodDelete, url, nil)

	accessToken := fmt.Sprintf("Bearer %s", azureAccessToken)

	req.Header.Add("Authorization", accessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil && (res.StatusCode == 400 || res.StatusCode == 404) {
		return uint16(res.StatusCode), err
	}
	defer res.Body.Close()

	if res.StatusCode == 204 {
		return uint16(res.StatusCode), nil
	} else {
		return 0, err
	}
}

func DeleteOAthApps(azureAccessToken, userPrincipalName string, ctx context.Context) (uint16, error) {
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/authentication/softwareOathMethods", userPrincipalName)

	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	accessToken := fmt.Sprintf("Bearer %s", azureAccessToken)
	req.Header.Add("Authorization", accessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil && (res.StatusCode == 400 || res.StatusCode == 404) {
		return uint16(res.StatusCode), err
	}
	defer res.Body.Close()

	if res.StatusCode == 204 {
		return uint16(res.StatusCode), nil
	} else {
		return 0, err
	}
}

func DeleteEmailAuthenticator(azureAccessToken, userPrincipalName string, ctx context.Context) (uint16, error) {
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/authentication/emailMethods/3ddfcfc8-9383-446f-83cc-3ab9be4be18f", userPrincipalName)

	req, _ := http.NewRequest(http.MethodDelete, url, nil)

	accessToken := fmt.Sprintf("Bearer %s", azureAccessToken)
	req.Header.Add("Authorization", accessToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil && (res.StatusCode == 400 || res.StatusCode == 404) {
		return uint16(res.StatusCode), err
	}
	defer res.Body.Close()

	if res.StatusCode == 204 {
		return uint16(res.StatusCode), nil
	} else {
		return 0, err
	}
}
