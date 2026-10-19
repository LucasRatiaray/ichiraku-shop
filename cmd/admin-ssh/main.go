// lance le serveur SSH exposant la TUI admin
// lance un serveur SSH (wish) exposant la même TUI client
package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/ssh"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"
	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
	"github.com/LucasRatiaray/ichiraku-shop/internal/tui/admin"
)

func main() {
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}
	api := apiclient.New(apiURL)

	teaHandler := func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		return admin.New(api), nil
	}

	srv, err := wish.NewServer(
		wish.WithAddress(":23235"),
		wish.WithHostKeyPath(".ssh/admin_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("SSH server listening on :23234")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
