package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"

	"github.com/Chandra5468/azure-ad-golang/helpers"
	"github.com/Chandra5468/azure-ad-golang/models/mango/tenants"
	"github.com/Chandra5468/azure-ad-golang/models/redis"
	"github.com/Chandra5468/azure-ad-golang/services"
)

type BodyCapture struct {
	TenantId          string `json:"x-tenant-id"`
	UserprincipleName string `json:"userPrincipalName"`
}
type AuthInfo struct {
	AuthName string
	Data     string
	Err      error
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

	// Go routine callings- Total 4 go routines
	// Calling Authententicators api here. using go routines for faster processing

	// Handling phone authenticators
	var wg *sync.WaitGroup
	wg.Add(4)
	resultChan := make(chan AuthInfo, 4)
	if lbl.Labelling.MfaMobileNumber != "" {
		go func() {
			defer wg.Done()
			data, err := services.GetPhoneAuthenticatorInfo(details.UserprincipleName, "3179e48a-750b-4051-897c-87b9720928f7", azureAccessToken)
			resultChan <- AuthInfo{
				AuthName: lbl.Labelling.MfaMobileNumber,
				Data:     data,
				Err:      err,
			}
		}()
	}
	if lbl.Labelling.MfaAlternativeMobileNumber != "" {
		go func() {
			defer wg.Done()
			data, err := services.GetPhoneAuthenticatorInfo(details.UserprincipleName, "b6332ec1-7057-4abe-9331-3d72feddfe41", azureAccessToken)
			resultChan <- AuthInfo{
				AuthName: lbl.Labelling.MfaAlternativeMobileNumber,
				Data:     data,
				Err:      err,
			}
		}()
	}
	if lbl.Labelling.MfaOfficeMobileNumber != "" {
		go func() {
			defer wg.Done()
			data, err := services.GetPhoneAuthenticatorInfo(details.UserprincipleName, "e37fc753-ff3b-4958-9484-eaa9425c82bc", azureAccessToken)
			resultChan <- AuthInfo{
				AuthName: lbl.Labelling.MfaOfficeMobileNumber,
				Data:     data,
				Err:      err,
			}
		}()
	}

	// Handling MS Authenticator and getting device information

	if lbl.Labelling.MicrosoftAuthenticatorApp != "" {
		go func() {
			defer wg.Done()
			data, err := services.MicrosoftAuthDevice(details.UserprincipleName, azureAccessToken)
			resultChan <- AuthInfo{
				AuthName: lbl.Labelling.MicrosoftAuthenticatorApp,
				Data:     data,
				Err:      err,
			}
		}()
	}

	wg.Wait()

	for x := range resultChan {
		if x.Err == nil {
			switch x.AuthName {
			case "mobile":
				dataSent[lbl.Labelling.MfaMobileNumber] = x.Data
			case "alternative mobile":
				dataSent[lbl.Labelling.MfaAlternativeMobileNumber] = x.Data
			case "office no":
				dataSent[lbl.Labelling.MfaOfficeMobileNumber] = x.Data
			case "MSAuth":
				dataSent[lbl.Labelling.MicrosoftAuthenticatorApp] = x.Data
			}
		} else {
			dataSent[x.AuthName] = "NA"
		}
	}

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
