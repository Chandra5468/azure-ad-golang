package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Chandra5468/azure-ad-golang/helpers"
	"github.com/Chandra5468/azure-ad-golang/services"
)

type BodyCapture struct {
	TenantId          string `json:"x-tenant-id"`
	UserprincipleName string `json:"userPrincipalName"`
}

func GetMicrosoftAuthenticatorApp(w http.ResponseWriter, r *http.Request) {
	// headers := r.Header

	var details BodyCapture
	err := json.NewDecoder(r.Body).Decode(&details)
	tenantId := r.Header.Get("x-tenant-id")
	if tenantId == "" {
		tenantId = details.TenantId
	}
	azureAccessToken := "azureAccessToken_" + tenantId

	if err != nil {
		// Take response from helper to send error message
		helpers.ErrorFormatter(w, http.StatusBadRequest, err)
		return
	}

	dataInBytes, err := services.GetMicrosoftAuthenticatorApp(details.UserprincipleName, azureAccessToken)

	if err != nil {
		helpers.ErrorFormatter(w, http.StatusInternalServerError, err)
	}

	helpers.ResponseFormatter(w, 202, dataInBytes)
}
