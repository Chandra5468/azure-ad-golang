package controllers

import (
	"encoding/json"
	"net/http"
)

type Details struct {
	TenantId          string `json:"x-tenant-id"`
	UserprincipleName string `json:"userPrincipalName"`
}

func GetMicrosoftAuthenticatorApp(w http.ResponseWriter, r *http.Request) {
	var details Details
	err := json.NewDecoder(r.Body).Decode(&details)

	if err != nil {
		// Take response from helper to send error message
		return
	}
}
