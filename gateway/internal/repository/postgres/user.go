package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

const uniqueViolation = "23505"

const userColumns = `id, email, password_hash, locale, dir_pref, status, created_at`

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) ByEmail(ctx context.Context, email string) (model.User, error) {
	return r.one(ctx, `select `+userColumns+` from users where email = $1`, email)
}

func (r *UserRepository) ByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	return r.one(ctx, `select `+userColumns+` from users where id = $1`, id)
}

func (r *UserRepository) one(ctx context.Context, query string, args ...any) (model.User, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return model.User{}, err
	}

	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.User])
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, constants.ErrNotFound
	}
	return user, err
}

// Create claims the invite and inserts the user in one transaction. The
// conditional update is what stops two simultaneous registrations consuming
// the same code.
func (r *UserRepository) Create(ctx context.Context, user model.User, invite string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// The user must exist before the invite can reference it: used_by is a
	// foreign key. The transaction discards this insert if the claim fails.
	_, err = tx.Exec(ctx,
		`insert into users (id, email, password_hash, locale, dir_pref, status)
		 values ($1, $2, $3, $4, $5, $6)`,
		user.ID, user.Email, user.PasswordHash, user.Locale, user.DirPref, user.Status)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return constants.ErrConflict
		}
		return err
	}

	claimed, err := tx.Exec(ctx,
		`update invites set used_by = $1, used_at = now()
		  where code = $2 and used_by is null`, user.ID, invite)
	if err != nil {
		return err
	}
	if claimed.RowsAffected() == 0 {
		return constants.ErrUnauthorized
	}

	return tx.Commit(ctx)
}
