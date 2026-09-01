package repo

import (
	"context"
	"fmt"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/core"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type repositoryRegistry struct {
	auth   core.AuthRepository
	prof   core.ProfileRepository
	org    core.OrgRepository
	outbox core.OutboxRepository
}

func (r *repositoryRegistry) Auth() core.AuthRepository       { return r.auth }
func (r *repositoryRegistry) Profile() core.ProfileRepository { return r.prof }
func (r *repositoryRegistry) Org() core.OrgRepository         { return r.org }
func (r *repositoryRegistry) Outbox() core.OutboxRepository   { return r.outbox }

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) WithinTx(ctx context.Context, fn func(r core.Repositories) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	reg := &repositoryRegistry{
		auth:   NewAuthRepo(tx),
		prof:   NewProfileRepo(tx),
		org:    NewOrgRepo(tx),
		outbox: NewOutboxRepo(tx),
	}

	if err := fn(reg); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	return nil
}
