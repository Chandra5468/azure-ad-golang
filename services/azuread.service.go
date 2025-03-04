package services

import (
	"fmt"
	"io"
	"net/http"
)

func GetMicrosoftAuthenticatorApp(userPrincipalName, azureAccessToken string) ([]byte, error) {
	// options := make(map[string]string)
	// options["method"] = "GET"

	// json.Marshal(map[string]string{
	// 	""
	// }) GET request does not need body

	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/authentication/microsoftAuthenticatorMethods", userPrincipalName)

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		// Here handle an error sending message
	}
	accessToken := fmt.Sprintf("Bearer %v", azureAccessToken)
	req.Header.Add("Authorization", accessToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)

	if err != nil {

	}

	defer res.Body.Close()

	byteData, err := io.ReadAll(res.Body)

	// res.Body.Read() Use this if the response is too large
	/* IF Response is too large like video data or file reading use below code
		for {
	        n, err := resp.Body.Read(buffer)
	        if err != nil {
	            if err == io.EOF {
	                break // End of stream
	            }
	            return err
	        }

	        _, err = file.Write(buffer[:n]) // Write the chunk to the file
	        if err != nil {
	            return err
	        }
	    }
	*/

	if err != nil {

	}

	return byteData, nil
}
