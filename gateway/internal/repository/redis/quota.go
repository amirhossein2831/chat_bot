package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// reserveScript compares spend plus outstanding holds against the ceiling
// atomically. Doing this in application code lets concurrent requests each
// pass the check and collectively overshoot, which is exactly the failure that
// turns a free tier into an unbounded bill.
var reserveScript = redis.NewScript(`
local spent = tonumber(redis.call('GET', KEYS[1]) or '0')
local held = 0
for _, v in ipairs(redis.call('HVALS', KEYS[2])) do
  held = held + tonumber(v)
end

if spent + held + tonumber(ARGV[2]) > tonumber(ARGV[1]) then
  return -1
end

redis.call('HSET', KEYS[2], ARGV[3], ARGV[2])
redis.call('EXPIRE', KEYS[2], tonumber(ARGV[4]))
return 1
`)

const (
	holdTTL     = 30 * time.Minute
	counterTTL  = 32 * 24 * time.Hour
	spentPrefix = "quota:spent:"
	holdsPrefix = "quota:holds:"
	countPrefix = "quota:count:"
)

type QuotaRepository struct {
	client *redis.Client
}

func NewQuotaRepository(client *redis.Client) *QuotaRepository {
	return &QuotaRepository{client: client}
}

// Reserve holds the worst-case cost of a request. It returns false when the
// hold would breach the ceiling.
func (r *QuotaRepository) Reserve(ctx context.Context, userID uuid.UUID, holdID string, amount, ceiling int64) (bool, error) {
	keys := []string{spentKey(userID), holdsKey(userID)}
	args := []any{ceiling, amount, holdID, int(holdTTL.Seconds())}

	result, err := reserveScript.Run(ctx, r.client, keys, args...).Int64()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

// Settle releases the hold and records what was actually spent.
func (r *QuotaRepository) Settle(ctx context.Context, userID uuid.UUID, holdID string, actual int64) error {
	pipe := r.client.TxPipeline()
	pipe.HDel(ctx, holdsKey(userID), holdID)
	if actual > 0 {
		pipe.IncrBy(ctx, spentKey(userID), actual)
		pipe.Expire(ctx, spentKey(userID), counterTTL)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *QuotaRepository) Release(ctx context.Context, userID uuid.UUID, holdID string) error {
	return r.client.HDel(ctx, holdsKey(userID), holdID).Err()
}

func (r *QuotaRepository) IncrementDaily(ctx context.Context, userID uuid.UUID) (int64, error) {
	key := dailyKey(userID)
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		_ = r.client.Expire(ctx, key, 48*time.Hour).Err()
	}
	return count, nil
}

func (r *QuotaRepository) DailyCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.client.Get(ctx, dailyKey(userID)).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

func spentKey(userID uuid.UUID) string {
	return fmt.Sprintf("%s%s:%s", spentPrefix, userID, time.Now().UTC().Format("2006-01"))
}

func holdsKey(userID uuid.UUID) string { return holdsPrefix + userID.String() }

func dailyKey(userID uuid.UUID) string {
	return fmt.Sprintf("%s%s:%s", countPrefix, userID, time.Now().UTC().Format("2006-01-02"))
}
