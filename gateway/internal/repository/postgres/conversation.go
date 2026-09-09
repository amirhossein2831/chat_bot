package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
	"github.com/amirhossein2831/chat_bot/gateway/internal/domain"
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
	"github.com/amirhossein2831/chat_bot/gateway/internal/util/textnorm"
)

const conversationColumns = `id, user_id, title, title_norm, default_model,
	head_message_id, archived_at, created_at`

const messageColumns = `id, conversation_id, parent_id, role, content, content_norm,
	status, model, tokens_in, tokens_out, cost_usd_micros, error, created_at`

// threadQuery walks parent pointers upward from the conversation's head, so
// the sibling branches left behind by regeneration are excluded by
// construction rather than by filtering.
const threadWalk = `
  union all
    select p.* from messages p
      join path on path.parent_id = p.id
)
select ` + messageColumns + ` from path limit $2`

const threadQuery = `
with recursive path as (
    select m.* from messages m
      join conversations c on c.head_message_id = m.id
     where c.id = $1` + threadWalk

// threadToQuery walks from a named message instead of the head, which is what
// a regeneration needs: the branch being replaced is not on the path.
const threadToQuery = `
with recursive path as (
    select m.* from messages m where m.id = $1` + threadWalk

type ConversationRepository struct {
	db *pgxpool.Pool
}

func NewConversationRepository(db *pgxpool.Pool) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// Create returns the stored row so created_at comes from the database rather
// than from a zero value the caller never set.
func (r *ConversationRepository) Create(ctx context.Context, c model.Conversation) (model.Conversation, error) {
	rows, err := r.db.Query(ctx,
		`insert into conversations (id, user_id, title, title_norm, default_model)
		 values ($1, $2, $3, $4, $5)
		 returning `+conversationColumns,
		c.ID, c.UserID, c.Title, textnorm.Normalize(c.Title), c.DefaultModel)
	if err != nil {
		return model.Conversation{}, err
	}
	return pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.Conversation])
}

// Get is scoped by user in the query itself. Ownership checked in a service is
// an IDOR waiting to happen; a missing join here simply returns no rows.
func (r *ConversationRepository) Get(ctx context.Context, id, userID uuid.UUID) (model.Conversation, error) {
	rows, err := r.db.Query(ctx,
		`select `+conversationColumns+` from conversations where id = $1 and user_id = $2`,
		id, userID)
	if err != nil {
		return model.Conversation{}, err
	}

	conversation, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.Conversation])
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Conversation{}, constants.ErrNotFound
	}
	return conversation, err
}

func (r *ConversationRepository) List(ctx context.Context, userID uuid.UUID, limit int) ([]model.Conversation, error) {
	rows, err := r.db.Query(ctx,
		`select `+conversationColumns+`
		   from conversations
		  where user_id = $1 and archived_at is null
		  order by created_at desc
		  limit $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[model.Conversation])
}

func (r *ConversationRepository) Search(ctx context.Context, userID uuid.UUID, query string, limit int) ([]model.Conversation, error) {
	normalized := textnorm.Normalize(query)
	if normalized == "" {
		return nil, nil
	}

	rows, err := r.db.Query(ctx,
		`select `+conversationColumns+`
		   from conversations c
		  where c.user_id = $1
		    and (c.title_norm like '%' || $2 || '%'
		         or exists (select 1 from messages m
		                     where m.conversation_id = c.id
		                       and m.content_norm like '%' || $2 || '%'))
		  order by c.created_at desc
		  limit $3`, userID, normalized, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[model.Conversation])
}

func (r *ConversationRepository) Rename(ctx context.Context, id, userID uuid.UUID, title string) error {
	tag, err := r.db.Exec(ctx,
		`update conversations set title = $3, title_norm = $4
		  where id = $1 and user_id = $2`,
		id, userID, title, textnorm.Normalize(title))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *ConversationRepository) Archive(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.db.Exec(ctx,
		`update conversations set archived_at = now()
		  where id = $1 and user_id = $2 and archived_at is null`,
		id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *ConversationRepository) Thread(ctx context.Context, conversationID uuid.UUID, limit int) (domain.Thread, error) {
	rows, err := r.db.Query(ctx, threadQuery, conversationID, limit)
	if err != nil {
		return domain.Thread{}, err
	}

	messages, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Message])
	if err != nil {
		return domain.Thread{}, err
	}
	return domain.ThreadFrom(messages), nil
}

func (r *ConversationRepository) ThreadTo(ctx context.Context, messageID uuid.UUID, limit int) (domain.Thread, error) {
	rows, err := r.db.Query(ctx, threadToQuery, messageID, limit)
	if err != nil {
		return domain.Thread{}, err
	}

	messages, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Message])
	if err != nil {
		return domain.Thread{}, err
	}
	return domain.ThreadFrom(messages), nil
}

func (r *ConversationRepository) Message(ctx context.Context, id uuid.UUID) (model.Message, error) {
	rows, err := r.db.Query(ctx, `select `+messageColumns+` from messages where id = $1`, id)
	if err != nil {
		return model.Message{}, err
	}

	m, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.Message])
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Message{}, constants.ErrNotFound
	}
	return m, err
}

func (r *ConversationRepository) Siblings(ctx context.Context, parentID uuid.UUID) ([]model.Message, error) {
	rows, err := r.db.Query(ctx,
		`select `+messageColumns+` from messages where parent_id = $1 order by created_at`,
		parentID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[model.Message])
}

func (r *ConversationRepository) Append(ctx context.Context, m model.Message) error {
	_, err := r.db.Exec(ctx,
		`insert into messages (id, conversation_id, parent_id, role, content, content_norm,
		                       status, model, tokens_in, tokens_out, cost_usd_micros)
		 values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		m.ID, m.ConversationID, m.ParentID, m.Role, m.Content, textnorm.Normalize(m.Content),
		m.Status, m.Model, m.TokensIn, m.TokensOut, m.CostUSDMicros)
	return err
}

// AppendPair inserts the user turn and the assistant placeholder together and
// moves the head, so a conversation is never left pointing at a message whose
// sibling was written but not linked.
func (r *ConversationRepository) AppendPair(ctx context.Context, user, assistant model.Message) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, m := range []model.Message{user, assistant} {
		if err := insertMessage(ctx, tx, m); err != nil {
			return err
		}
	}

	if err := moveHead(ctx, tx, assistant); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// AppendBranch adds one assistant message beside an existing sibling and moves
// the head onto it, which is how a regenerated answer replaces the one before
// it without destroying it.
func (r *ConversationRepository) AppendBranch(ctx context.Context, m model.Message) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := insertMessage(ctx, tx, m); err != nil {
		return err
	}
	if err := moveHead(ctx, tx, m); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func insertMessage(ctx context.Context, tx pgx.Tx, m model.Message) error {
	_, err := tx.Exec(ctx,
		`insert into messages (id, conversation_id, parent_id, role, content, content_norm,
		                       status, model, tokens_in, tokens_out, cost_usd_micros)
		 values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		m.ID, m.ConversationID, m.ParentID, m.Role, m.Content, textnorm.Normalize(m.Content),
		m.Status, m.Model, m.TokensIn, m.TokensOut, m.CostUSDMicros)
	return err
}

func moveHead(ctx context.Context, tx pgx.Tx, m model.Message) error {
	_, err := tx.Exec(ctx,
		`update conversations set head_message_id = $2 where id = $1`,
		m.ConversationID, m.ID)
	return err
}

func (r *ConversationRepository) AppendPartial(ctx context.Context, id uuid.UUID, content string) error {
	_, err := r.db.Exec(ctx, `update messages set content = $2 where id = $1`, id, content)
	return err
}

func (r *ConversationRepository) Finalize(ctx context.Context, m model.Message) error {
	_, err := r.db.Exec(ctx,
		`update messages
		    set content = $2, content_norm = $3, status = $4,
		        tokens_in = $5, tokens_out = $6, cost_usd_micros = $7, error = $8
		  where id = $1`,
		m.ID, m.Content, textnorm.Normalize(m.Content), m.Status,
		m.TokensIn, m.TokensOut, m.CostUSDMicros, m.Error)
	return err
}

func (r *ConversationRepository) SetHead(ctx context.Context, conversationID, messageID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`update conversations set head_message_id = $2 where id = $1`,
		conversationID, messageID)
	return err
}
