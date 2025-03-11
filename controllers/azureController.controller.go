package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Chandra5468/azure-ad-golang/helpers"
	"github.com/Chandra5468/azure-ad-golang/models/mango/tenants"
	"github.com/Chandra5468/azure-ad-golang/models/redis"
	"github.com/Chandra5468/azure-ad-golang/services"
)

type BodyCapture struct {
	TenantId          string `json:"x-tenant-id"`
	UserprincipleName string `json:"userPrincipalName"`
}

func GetUserAllInfo(w http.ResponseWriter, r *http.Request) {
	var details BodyCapture
	// Not using firstName and lastName based search

	err := json.NewDecoder(r.Body).Decode(&details)

	if err != nil {
		helpers.ErrorFormatter(w, http.StatusBadRequest, errors.New("unable to decode body. Send correct fields"))
		return
	}
	if details.UserprincipleName == "" {
		helpers.ErrorFormatter(w, http.StatusBadGateway, errors.New("userPricipal name not sent"))
		return
	}

	tenantId := r.Header.Get("x-tenant-id")

	if tenantId == "" {
		tenantId = details.TenantId
	}

	azureAccessToken, err := redis.CacheRead(r.Context(), "azureAccessToken_"+tenantId)
	if err != nil {
		helpers.ErrorFormatter(w, http.StatusInternalServerError, errors.New("unable to get access token from redis"))
		return
	}
	// Fetch what all values you want from azure. I.e.,, all the fields from mongo
	var lbl tenants.LabelConfigs
	labelConfigs, _ := redis.CacheRead(r.Context(), "azure_ad_labelling_"+tenantId)
	if labelConfigs == "" {
		data, err := tenants.GetAzureLabelConfigs(tenantId, r.Context())
		if err != nil {
			helpers.ErrorFormatter(w, http.StatusInternalServerError, errors.New("mongo error while getting label configs"))
			return
		}
		b, err := json.Marshal(&data)

		if err != nil {
			helpers.ErrorFormatter(w, http.StatusInternalServerError, errors.New("marshalling error of azure labelling"))
			return
		}
		err = redis.CacheWrite(r.Context(), "azure_ad_labelling_"+tenantId, string(b))
		if err != nil {
			helpers.ErrorFormatter(w, http.StatusInternalServerError, errors.New("error while writing labelling to cache"))
			return
		}
	} else {
		json.Unmarshal([]byte(labelConfigs), &lbl) // converting string to struct for better usage convinience
	}

	userInfo, err := services.GetUserInfo(details.UserprincipleName, azureAccessToken)
	if err != nil {
		helpers.ErrorFormatter(w, http.StatusInternalServerError, err)
		return
	}
	dataSent := make(map[string]string) // This will be sent with all the data

	dataSent[lbl.Labelling.UserprincipleName] = userInfo.UserprincipleName
	dataSent[lbl.Labelling.DisplayName] = userInfo.DisplayName
	dataSent[lbl.Labelling.AccountEnabled] = strconv.FormatBool(userInfo.AccountEnabled)
	// dataSent[lbl.Labelling.BusinessPhones] = userInfo.BusinessPhones
	dataSent[lbl.Labelling.LastPWDChangeDateTime] = userInfo.LastPWDChangeDateTime
	// Calling Authententicators api here. using go routines for faster processing
	// if dataSent[lbl.la]
	helpers.ResponseFormatter(w, http.StatusOK, &dataSent)
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
