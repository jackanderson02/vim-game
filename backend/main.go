package main

import (
	"log"
	"fmt"
	"net/http"
	"github.com/neovim/go-client/nvim"
)

type VimInstance struct {
	vim *nvim.Nvim
	window nvim.Window
	cursor Cursor
}

type Cursor struct {
	Row int // 1 indexed `json: "row" vim:"row"`
	Column int // 0 indexed `json: "column" vim:"column"`
}

var vi VimInstance;

func CorsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
        writer.Header().Set("Access-Control-Allow-Origin", "*")
        // writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        // writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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
	log.Println("Initializing Neovim...")
	vim, cleanup, err := connectWhenReady()
	if err != nil {
		log.Fatalf("Failed to initialize Neovim: %v", err)
	}
	defer cleanup()


	windows, err := vim.Windows()
	if err != nil {
		log.Fatalf("Failed to get windows %v", err)
	}
	if len(windows) == 0 {
		log.Fatal("No windows found.")
	}

	vi = VimInstance{
		vim,
		windows[0],
		Cursor{
			1, 0,
		},
		
	}


	log.Println("Creating a new buffer...")
	buffer, err := vim.CreateBuffer(true, true)
	if err != nil {
		log.Fatalf("Failed to create buffer: %v", err)
	}

	log.Println("Setting text content in the buffer...")
	text := [][]byte{
		[]byte("Hello, world!"),
		[]byte("This is a test."),
		[]byte("Vim keybindings in Go."),
	}

	if err := vim.SetBufferLines(buffer, 0, -1, true, text); err != nil {
		log.Fatalf("Failed to set buffer lines: %v", err)
	}

	log.Println("Attaching the buffer to the current window...")
	if err := vim.SetCurrentBuffer(buffer); err != nil {
		log.Fatalf("Failed to set current buffer: %v", err)
	}

	router := http.NewServeMux()

	router.HandleFunc("GET /keyPress", vi.handleKeyPress)
	// router.HandleFunc("GET /api/users/{id}", userAPI.GetSingleUser)
	// router.HandleFunc("GET /api/users", userAPI.GetUsers)
	// router.HandleFunc("DELETE /api/users/{id}", userAPI.DeleteUser)
	// router.HandleFunc("PUT /api/users/{id}", userAPI.UpdateUser)
	// router.HandleFunc("POST /api/users", userAPI.CreateUser)
	// router.HandleFunc("GET /api/products", productAPI.GetProducts)

	// Starting the HTTP server on port 8080
	fmt.Println("Server listening on port 8080...")
	err = http.ListenAndServe(":8080", CorsMiddleware(router))
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
