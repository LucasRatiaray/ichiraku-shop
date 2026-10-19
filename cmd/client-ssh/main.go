// lance un serveur SSH (wish) exposant la même TUI client
package main

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"charm.land/ssh"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"
	"github.com/LucasRatiaray/ichiraku-shop/internal/tui/client"
)

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	return client.New(), nil
}

func main() {
	server, err := wish.NewServer(
		wish.WithAddress(":23234"),
		wish.WithHostKeyPath(".ssh/client_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("SSH server listening on :23234")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}