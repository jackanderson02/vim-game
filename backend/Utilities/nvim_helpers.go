package util 

import (
	"fmt"
	"log"
	"time"
	"github.com/neovim/go-client/nvim"
)
// Connect to Neovim when it's ready

func ConnectWhenReady() (*nvim.Nvim, func(), error) {
	const maxRetries = 10
	const waitDuration = 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		log.Println("Attempting to connect to Neovim...")

		vim, err := nvim.Dial("127.0.0.1:6665")
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

