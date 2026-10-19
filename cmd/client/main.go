// lance la TUI client en terminal local
package main

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/LucasRatiaray/ichiraku-shop/internal/tui/client"
)

func main() {
	if _, err := tea.NewProgram(client.New()).Run(); err != nil {
		log.Fatal(err)
	}
}
