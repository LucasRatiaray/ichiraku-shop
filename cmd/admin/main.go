// lance la TUI admin en terminal local
package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
	"github.com/LucasRatiaray/ichiraku-shop/internal/tui/admin"
)

func main() {
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}
	api := apiclient.New(apiURL)

	if _, err := tea.NewProgram(admin.New(api)).Run(); err != nil {
		log.Fatalf("admin TUI: %v", err)
	}
}
