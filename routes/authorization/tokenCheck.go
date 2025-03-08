package authorization

import (
	"errors"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Chandra5468/azure-ad-golang/helpers"
	"github.com/Chandra5468/azure-ad-golang/models/mango/tenants"
	"github.com/Chandra5468/azure-ad-golang/models/redis"
	"github.com/Chandra5468/azure-ad-golang/services"
)

// func CheckCredentials(next http.Handler) http.HandlerFunc {
func CheckCredentials(next http.HandlerFunc) http.HandlerFunc {
	// http.Handle() //
	// http.HandleFunc() // Direct method route mentioning and handler mentioning. Mostly used for simple application direct routing
	// http.Handler // Interface, any struct of our own type which is assigned to this interface can implement this.
	// http.HandlerFunc // This is a type of HandleFunc() ex: func(ResponseWriter, *Request)

	// IMP : **** func -> HandlerFunc -> Handler

	// Explaination of each
	/*
		http.Handler : This implements interface
		we cannot pass a HandlerFunc to as a http.Handler. This will not implement interface

		***** http.HandlerFunc : This is a type. This is type of func(ResponseWriter, *Request)
		So when you implement a middleware.
		in route(middle1(controller))
				in controller |
		         middle1(next http.Handler)

		----This above will give you an error.
		in route(middle1(http.HandlerFunc(controller)))
				in controller |
				middle1(next http.Handler)

		--- This will not give an error, as we are implementing http.Handler interface by passing type HandlerFunc in it.
	*/
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantId := r.Header.Get("x-tenant-id")
		accessTokenKey := "azureAccessToken_" + tenantId
		azureAccessTokenString, err := redis.CacheRead(r.Context(), accessTokenKey)
		if err != nil {
			slog.Error("error while reading from redis_", "error", err.Error())
			helpers.ErrorFormatter(w, http.StatusInternalServerError, err)
			return
		}
		keyExpiry, err := redis.CacheKeyTTL(r.Context(), accessTokenKey)
		if err != nil { // comment this and use _ for above err
			log.Println("key timeout error", keyExpiry)
		}
		if keyExpiry*time.Second < 300 {
			productDetails, err := tenants.GetAzureConfigs(tenantId, r.Context())
			if err != nil {
				helpers.ErrorFormatter(w, http.StatusInternalServerError, errors.New("not able to get azure configurations from mongo"))
				return
			}
			if productDetails.Products.AzureActiveDirectory.GrantType == "password" {
				accessToken, err := services.GetAccessTokenPGgrant(&productDetails.Products.AzureActiveDirectory)
				if err != nil {
					helpers.ErrorFormatter(w, http.StatusInternalServerError, errors.New("issue while generating access token from azure end, please try after sometime"))
					return
				} else {
					intTime, _ := strconv.Atoi(accessToken.ExpiresIn)
					redis.CacheWriteWithExpiry(r.Context(), accessTokenKey, accessToken.AccessToken, time.Duration(time.Duration(intTime).Seconds()))
					next.ServeHTTP(w, r)
				}
			} else {
				helpers.ErrorFormatter(w, http.StatusBadRequest, errors.New("out of scope azure connectivity"))
				return
			}
		} else if azureAccessTokenString != "" {
			next.ServeHTTP(w, r)
		} else {
			helpers.ErrorFormatter(w, http.StatusInternalServerError, errors.New("unable to get any information on access token"))
			return
		}
	})
}
