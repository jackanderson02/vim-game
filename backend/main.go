package main

import (
	"fmt"
	"net/http"
	"vim-zombies/Auth"
)

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle OPTIONS method (preflight request)
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusOK)
			return
		}

		// Continue with the next handler
		next.ServeHTTP(writer, request)
	})
}

func main() {

	// Now, just get new play Instance
	// gameInstance := game.NewInstance()

	router := http.NewServeMux()
	// Add authentication layer to each endpoint
	router.HandleFunc("POST /level", auth.GetLevelWrapper)
	router.HandleFunc("POST /resetLevel", auth.ResetLevelWrapper)
	router.HandleFunc("POST /keyPress", auth.HandleKeyPressWrapper)
	// router.HandleFunc("GET /level/", gameInstance.GetLevel)
	// router.HandleFunc("GET /resetLevel", gameInstance.ResetLevel)
	// router.HandleFunc("POST /keyPress", gameInstance.HandleKeyPress)

	// router.HandleFunc("DELETE /api/users/{id}", userAPI.DeleteUser)
	// router.HandleFunc("PUT /api/users/{id}", userAPI.UpdateUser)
	// Starting the HTTP server on port 8080
	fmt.Println("Server listening on port 8080...")
	err := http.ListenAndServe(":8080", CorsMiddleware(router))
	if err != nil {
		fmt.Println("Error starting server:", err)
	}


	// defer gameInstance.Cleanup()
}
