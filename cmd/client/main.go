// lance la TUI client en terminal local
package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
	"github.com/LucasRatiaray/ichiraku-shop/internal/tui/client"
)

func main() {
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}
	api := apiclient.New(apiURL)

	if _, err := tea.NewProgram(client.New(api)).Run(); err != nil {
		log.Fatal(err)
	}
}
