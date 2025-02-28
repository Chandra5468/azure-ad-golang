package v1

import (
	"net/http"

	"github.com/Chandra5468/azure-ad-golang/controllers"
)

func AzureRoutes() {
	router := http.NewServeMux()

	router.HandleFunc("POST /azuread/get/microsoftAuthenticator", controllers.GetMicrosoftAuthenticatorApp)
	// router.Handler()
}
