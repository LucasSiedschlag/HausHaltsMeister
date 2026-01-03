package audit

import (
	"context"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	domain "github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/audit"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(store *postgres.Store) *Repository {
	return &Repository{pool: store.Pool()}
}

func (r *Repository) Record(ctx context.Context, event domain.Event) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_log (ledger_id, user_id, action, entity_id, ip, user_agent, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NULL)
	`, event.LedgerID, event.UserID, event.Action, event.EntityID, event.IP, event.UserAgent)
	return err
}
