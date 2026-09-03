package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type RepositoryRegistry struct {
	proj   *ProjectRepo
	task   *TaskRepo
	an     *AnalyticsRepo
	inbox  *InboxRepo
	outbox *OutboxRepo
	gitlab *GitLabRepo
}

func (r *RepositoryRegistry) Project() *ProjectRepo     { return r.proj }
func (r *RepositoryRegistry) Task() *TaskRepo           { return r.task }
func (r *RepositoryRegistry) Analytics() *AnalyticsRepo { return r.an }
func (r *RepositoryRegistry) Inbox() *InboxRepo         { return r.inbox }
func (r *RepositoryRegistry) Outbox() *OutboxRepo       { return r.outbox }
func (r *RepositoryRegistry) GitLab() *GitLabRepo       { return r.gitlab }

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) WithinTx(ctx context.Context, fn func(r *RepositoryRegistry) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	reg := &RepositoryRegistry{
		proj:   NewProjectRepo(tx),
		task:   NewTaskRepo(tx),
		an:     NewAnalyticsRepo(tx),
		inbox:  NewInboxRepo(tx),
		outbox: NewOutboxRepo(tx),
		gitlab: NewGitLabRepo(tx),
	}

	if err := fn(reg); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	return nil
}
