package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/neovim/go-client/nvim"
)

type KeyPress struct {
	Key string `json:"key" vim:"key"`
	// Open to add additional fields such as time stamp
}

var errorCursor Cursor = Cursor{
	-1, -1,
}

// Connect to Neovim when it's ready
func connectWhenReady() (*nvim.Nvim, func(), error) {
	const maxRetries = 10
	const waitDuration = 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		log.Println("Attempting to connect to Neovim...")

		vim, err := nvim.Dial("127.0.0.1:6666")
		if err == nil {
			log.Println("Connected to Neovim.")
			cleanup := func() {
				vim.Close()
			}
			return vim, cleanup, nil
		}
		time.Sleep(waitDuration) // Wait before retrying
	}

	return nil, nil, fmt.Errorf("failed to connect to Neovim after multiple retries")
}


// Function to get the cursor position
func (vi *VimInstance) getCursorPosition() (Cursor, error) {
	vim := vi.vim
	mode, _ := vim.Mode()
	if(mode.Blocking){
		return errorCursor, fmt.Errorf("nvim is currently blocked because it is waiting additional user input.")
	}

	cursor, err := vim.WindowCursor(vi.window)
	if err != nil {
		return errorCursor, fmt.Errorf("failed to get cursor position: %v", err)
	}

	return Cursor{
		cursor[0], cursor[1],
	}, nil 
}

func getLastErrorMessage(vim *nvim.Nvim) (string, error) {
	var errmsg string
	if err := vim.Eval("v:errmsg", &errmsg); err != nil {
		return "", fmt.Errorf("failed to get v:errmsg: %v", err)
	}
	return errmsg, nil
}


func (vi *VimInstance) handleKeyPress(writer http.ResponseWriter, request *http.Request){

	var keypress KeyPress
	// Decode keypress from json
	err := json.NewDecoder(request.Body).Decode(&keypress)

	if err != nil{
		log.Print("Error decoding message request body:")
		return
	}

	log.Printf("Got keypress %s.", keypress.Key)

	// Make the keypress
	vi.vim.Input(keypress.Key)

	cursorPosition, _ := vi.getCursorPosition()
	
	if cursorPosition == errorCursor{
		// Can handle this case if needed, but for now this just means incomplete input sequence

		// Current position
		log.Print("Current position")
		log.Print(vi.cursor)
		log.Print("Key press only forms a partial input sequence.")
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(vi.cursor)

	}else{
		// Do not change cursor if incomplete input sequence used
		vi.cursor = cursorPosition 
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(vi.cursor)
	}


	

}
