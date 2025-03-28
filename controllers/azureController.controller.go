package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Chandra5468/azure-ad-golang/helpers"
	"github.com/Chandra5468/azure-ad-golang/logging"
	"github.com/Chandra5468/azure-ad-golang/models/mango/tenants"
	"github.com/Chandra5468/azure-ad-golang/models/redis"
	"github.com/Chandra5468/azure-ad-golang/services"
)

type BodyCapture struct {
	TenantId          string `json:"x-tenant-id"`
	UserprincipleName string `json:"userPrincipalName"`
	NewPassword       string `json:"newPassword"`
}
type AuthInfo struct {
	AuthName string
	Data     string
	Err      error
}

type DeleteAuthInfo struct {
	Status   uint16
	Err      error
	AuthName string
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
	var wg sync.WaitGroup
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
	close(resultChan)

	for x := range resultChan {
		if x.Err == nil {
			switch x.AuthName {
			case lbl.Labelling.MfaMobileNumber:
				dataSent[lbl.Labelling.MfaMobileNumber] = x.Data
			case lbl.Labelling.MfaAlternativeMobileNumber:
				dataSent[lbl.Labelling.MfaAlternativeMobileNumber] = x.Data
			case lbl.Labelling.MfaOfficeMobileNumber:
				dataSent[lbl.Labelling.MfaOfficeMobileNumber] = x.Data
			case lbl.Labelling.MicrosoftAuthenticatorApp:
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

func ResetAzurePassword(w http.ResponseWriter, r *http.Request) {
	var details BodyCapture
	logEntry := &logging.AzureESOps{
		ActionPerformed:   "Reset Password go",
		ActionPerformedAt: time.Now(),
	}
	defer logging.LogIntoAzureIndex(logEntry)
	err := json.NewDecoder(r.Body).Decode(&details)

	if err != nil {
		logEntry.Error = err
		helpers.ErrorFormatter(w, http.StatusBadRequest, errors.New("unable to decode body"))
		return
	}
	tenantId := r.Header.Get("x-tenant-id")
	if tenantId == "" {
		tenantId = details.TenantId
	}
	logEntry.ClientName = tenantId
	logEntry.UserId = details.UserprincipleName
	azureAccessToken, err := redis.CacheRead(r.Context(), "azureAccessToken_"+tenantId)
	if err != nil {
		// Take response from helper to send error message
		logEntry.Error = err
		helpers.ErrorFormatter(w, http.StatusInternalServerError, errors.New("unable to get access token from redis"))
		return
	}

	statusCode, err := services.AzurePwdReset(details.UserprincipleName, azureAccessToken, details.NewPassword)

	if err != nil {
		logEntry.Error = err
		helpers.ErrorFormatter(w, http.StatusBadRequest, err)
		return
	}
	logEntry.Successfull = true
	helpers.ResponseFormatter(w, statusCode, "Password has been reset successfully")
}

func DeleteConfiguredAuthenticators(w http.ResponseWriter, r *http.Request) {
	var details BodyCapture
	logEntry := &logging.AzureESOps{
		ActionPerformed:   "Delete All Auths",
		ActionPerformedAt: time.Now(),
	}
	defer logging.LogIntoAzureIndex(logEntry)
	err := json.NewDecoder(r.Body).Decode(&details)

	if err != nil {
		logEntry.Error = err
		helpers.ErrorFormatter(w, http.StatusInternalServerError, err)
		return
	}

	tenantId := r.Header.Get("x-tenant-id")
	if tenantId == "" {
		tenantId = details.TenantId
	}
	logEntry.ClientName = tenantId
	logEntry.UserId = details.UserprincipleName
	azureAccessToken, err := redis.CacheRead(r.Context(), "azureAccessToken_"+tenantId)

	if err != nil {
		logEntry.Error = err
		helpers.ErrorFormatter(w, http.StatusInternalServerError, errors.New("unable to get access token from redis"))
		return
	}
	resultChan := make(chan DeleteAuthInfo, 4)
	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		status, err := services.DeleteEmailAuthenticator(azureAccessToken, details.UserprincipleName, r.Context())
		resultChan <- DeleteAuthInfo{
			Status:   status,
			Err:      err,
			AuthName: "Email",
		}
	}()
	go func() {
		defer wg.Done()
		status, err := services.DeleteMicrosoftAuthenticators(azureAccessToken, details.UserprincipleName, r.Context())
		resultChan <- DeleteAuthInfo{
			Status:   status,
			Err:      err,
			AuthName: "MSAuth",
		}
	}()
	go func() {
		defer wg.Done()
		status, err := services.DeleteOAthApps(azureAccessToken, details.UserprincipleName, r.Context())
		resultChan <- DeleteAuthInfo{
			Status:   status,
			Err:      err,
			AuthName: "OAthApps",
		}
	}()
	go func() {
		defer wg.Done()
		// currently only deleting 1 phone authenticator device
		status, err := services.DeletePhoneAuthenticators(azureAccessToken, details.UserprincipleName, r.Context())
		resultChan <- DeleteAuthInfo{
			Status:   status,
			Err:      err,
			AuthName: "Phones",
		}
	}()
	wg.Wait()
	close(resultChan)
	for x := range resultChan {
		if x.Status == 400 {
			switch x.AuthName {
			case "Email":
				services.DeleteEmailAuthenticator(azureAccessToken, details.UserprincipleName, r.Context())
			case "Phones":
				services.DeletePhoneAuthenticators(azureAccessToken, details.UserprincipleName, r.Context())
			case "OAthApps":
				services.DeleteOAthApps(azureAccessToken, details.UserprincipleName, r.Context())
			case "MSAuth":
				services.DeleteMicrosoftAuthenticators(azureAccessToken, details.UserprincipleName, r.Context())
			}
		}
	}
	logEntry.Successfull = true
	helpers.ResponseFormatter(w, http.StatusNoContent, "all authenticators are successfully")

}
