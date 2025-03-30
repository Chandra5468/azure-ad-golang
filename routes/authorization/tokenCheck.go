package authorization

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Chandra5468/azure-ad-golang/helpers"
	"github.com/Chandra5468/azure-ad-golang/models/mango/tenants"
	"github.com/Chandra5468/azure-ad-golang/models/redis"
	"github.com/Chandra5468/azure-ad-golang/services"
)

// func CheckCredentials(next http.Handler) http.HandlerFunc {
func CheckCredentials(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantId := r.Header.Get("x-tenant-id")
		accessTokenKey := "azureAccessToken_" + tenantId
		azureAccessTokenString, _ := redis.CacheRead(r.Context(), accessTokenKey)

		keyExpiry, _ := redis.CacheKeyTTL(r.Context(), accessTokenKey)

		if keyExpiry == -2 || keyExpiry.Seconds() < 300 || azureAccessTokenString == "" {
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
					err = redis.CacheWriteWithExpiry(r.Context(), accessTokenKey, accessToken.AccessToken, intTime)
					if err != nil {
						helpers.ErrorFormatter(w, http.StatusInternalServerError, errors.New("issue while setting access token to redis with expiry"))
						return
					}
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
