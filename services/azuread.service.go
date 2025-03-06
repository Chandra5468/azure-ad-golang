package services

import (
	"encoding/json"
	"fmt"
	"net/http"
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
