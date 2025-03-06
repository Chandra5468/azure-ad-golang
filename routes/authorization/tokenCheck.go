package authorization

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Chandra5468/azure-ad-golang/helpers"
	"github.com/Chandra5468/azure-ad-golang/models/redis"
)

// type AzureAccessTokenCheck struct {
// 	TokenType   string `json:"token_type"`
// 	Scope       string `json:"scope"`
// 	ExpiresIn   string `json:"expires_in"`
// 	AccessToken string `json:"access_token"`
// }

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
		// var token AzureAccessTokenCheck
		accessTokenKey := "azureAccessToken" + tenantId
		var token map[string]string
		azureAccessTokenString, err := redis.CacheRead(r.Context(), accessTokenKey)
		if err != nil {
			slog.Error("error while reading from redis_", "error", err.Error())
			helpers.ErrorFormatter(w, http.StatusInternalServerError, err)
		} else {
			// json.NewDecoder().Decode()
			json.Unmarshal([]byte(azureAccessTokenString), &token)
		}
		keyExpiry, _ := redis.CacheKeyTTL(r.Context(), accessTokenKey)
		accessToken, ok := token[accessTokenKey]
		if ok || keyExpiry*time.Second < 300 {

		} else if ok && accessToken != "" {
			next.ServeHTTP(w, r)
		} else {
			helpers.ErrorFormatter(w, http.StatusInternalServerError, errors.New("unable to get any information on access token"))
		}
	})
}
