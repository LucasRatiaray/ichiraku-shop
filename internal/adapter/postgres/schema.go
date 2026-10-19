package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "embed"
)

//go:embed schema.sql
var schema string

// Migrate crée les tables de la base de données si elles n'existent pas.
func Migrate(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}
