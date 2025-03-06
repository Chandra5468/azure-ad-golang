package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Chandra5468/azure-ad-golang/helpers"
	"github.com/Chandra5468/azure-ad-golang/models/redis"
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
	if err != nil {
		// Take response from helper to send error message
		helpers.ErrorFormatter(w, http.StatusBadRequest, errors.New("unable to decode body"))
		return
	}
	tenantId := r.Header.Get("x-tenant-id")
	if tenantId == "" {
		tenantId = details.TenantId
	}
	azureAccessToken, err := redis.CacheRead(r.Context(), "azureAccessToken_"+tenantId)
	if err != nil {
		// Take response from helper to send error message
		helpers.ErrorFormatter(w, http.StatusInternalServerError, errors.New("unable to get access token from redis"))
		return
	}
	data, err := services.GetMicrosoftAuthenticatorApp(details.UserprincipleName, azureAccessToken)

	if err != nil {
		helpers.ErrorFormatter(w, http.StatusInternalServerError, errors.New("there is some error while getting authentication device infomation"))
		return
	}

	helpers.ResponseFormatter(w, http.StatusOK, data)
}
