package v1

import (
	"net/http"

	"github.com/Chandra5468/azure-ad-golang/controllers"
	"github.com/Chandra5468/azure-ad-golang/routes/authorization"
)

func AzureRoutes(router *http.ServeMux) {
	// func AzureRoutes() {
	// router := http.NewServeMux()

	// router.HandleFunc("POST /azuread/get/microsoftAuthenticator", authorization.CheckCredentials(http.HandlerFunc(controllers.GetMicrosoftAuthenticatorApp)))
	router.HandleFunc("POST /azuread/get/microsoftAuthenticator", authorization.CheckCredentials(controllers.GetMicrosoftAuthenticatorApp))
	// mux.Handle("POST /azuread/get/microsoftAuthenticator", authorization.CheckCredentials(controllers.GetMicrosoftAuthenticatorApp))
}
