package authorization

import "net/http"

func CheckCredentials(next http.Handler) http.HandlerFunc {
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
		// tenantId := r.Header.Get("x-tenant-id")
		// azureAccessToken :=
		next.ServeHTTP(w, r)
	})
}
