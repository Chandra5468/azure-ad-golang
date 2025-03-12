package v1

import (
	"net/http"

	"github.com/Chandra5468/azure-ad-golang/controllers"
	"github.com/Chandra5468/azure-ad-golang/routes/authorization"
)

func AzureRoutes(router *http.ServeMux) {
	router.HandleFunc("POST /azuread/user/info", authorization.CheckCredentials(controllers.GetUserAllInfo))
	router.HandleFunc("POST /azuread/get/microsoftAuthenticator", authorization.CheckCredentials(controllers.GetMicrosoftAuthenticatorApp))
	router.HandleFunc("POST /azuread/reset/password", authorization.CheckCredentials(controllers.ResetAzurePassword))
}
