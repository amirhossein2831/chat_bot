package quota

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"
	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
	"github.com/amirhossein2831/chat_bot/gateway/internal/domain"
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

// Reservation is what a caller must settle. Holding one is the only thing that
// keeps concurrent requests from collectively overshooting a ceiling.
type Reservation struct {
	ID     string
	UserID uuid.UUID
	Model  string
	Hold   int64
}

type Service struct {
	counters Counters
	ledger   Ledger
	models   map[string]config.ModelConfig
	tiers    map[string]provider.Tier
}

func NewService(counters Counters, ledger Ledger, cfg *config.Config) *Service {
	models := make(map[string]config.ModelConfig, len(cfg.Models))
	tiers := make(map[string]provider.Tier, len(cfg.Models))
	for _, m := range cfg.Models {
		models[m.ID] = m
		tiers[m.ID] = provider.Tier(m.Tier)
	}
	return &Service{counters: counters, ledger: ledger, models: models, tiers: tiers}
}

func (s *Service) PlanFor(ctx context.Context, userID uuid.UUID) (Plan, error) {
	code, expiresAt, err := s.ledger.ActiveSubscription(ctx, userID)
	if err != nil {
		return Plan{}, err
	}
	if code == "" {
		return PlanFor(PlanFree), nil
	}

	subscription := Subscription{PlanCode: PlanCode(code), ExpiresAt: expiresAt}
	if !subscription.Active(time.Now()) {
		return PlanFor(PlanFree), nil
	}
	return PlanFor(subscription.PlanCode), nil
}

// Reserve holds the worst case: every prompt token plus the maximum the model
// could emit. Settle refunds the difference.
func (s *Service) Reserve(ctx context.Context, userID uuid.UUID, modelID string, promptTokens, maxTokens int) (Reservation, error) {
	plan, err := s.PlanFor(ctx, userID)
	if err != nil {
		return Reservation{}, err
	}

	if !plan.Allows(s.tiers[modelID]) {
		return Reservation{}, fmt.Errorf("%w: plan %s cannot use %s",
			constants.ErrModelUnavailable, plan.Code, modelID)
	}

	count, err := s.counters.DailyCount(ctx, userID)
	if err != nil {
		return Reservation{}, err
	}
	if count >= int64(plan.DailyMessages) {
		return Reservation{}, fmt.Errorf("%w: %d messages today", constants.ErrQuotaExceeded, count)
	}

	hold := domain.EstimateCost(s.models[modelID], promptTokens, maxTokens)
	reservation := Reservation{
		ID:     uuid.NewString(),
		UserID: userID,
		Model:  modelID,
		Hold:   hold,
	}

	// A zero-cost model still consumes the daily message allowance, but has no
	// ceiling to breach, so it skips the reservation entirely.
	if hold > 0 {
		ok, err := s.counters.Reserve(ctx, userID, reservation.ID, hold, plan.MonthlyCostMicros)
		if err != nil {
			return Reservation{}, err
		}
		if !ok {
			return Reservation{}, fmt.Errorf("%w: monthly ceiling reached", constants.ErrQuotaExceeded)
		}
	}

	if _, err := s.counters.IncrementDaily(ctx, userID); err != nil {
		return Reservation{}, err
	}

	return reservation, nil
}

// Settle records what the generation actually cost and releases the hold. The
// ledger write is the source of truth; the counter is a cache.
func (s *Service) Settle(ctx context.Context, r Reservation, messageID uuid.UUID, usage provider.Usage) (model.Usage, error) {
	actual := model.Usage{
		In:            usage.In,
		Out:           usage.Out,
		CostUSDMicros: domain.EstimateCost(s.models[r.Model], usage.In, usage.Out),
	}

	if err := s.counters.Settle(ctx, r.UserID, r.ID, actual.CostUSDMicros); err != nil {
		return actual, err
	}
	if actual.CostUSDMicros == 0 && actual.In == 0 && actual.Out == 0 {
		return actual, nil
	}
	return actual, s.ledger.Record(ctx, r.UserID, messageID, r.Model, actual)
}

func (s *Service) Release(ctx context.Context, r Reservation) error {
	return s.counters.Release(ctx, r.UserID, r.ID)
}

func (s *Service) CostOf(modelID string, usage provider.Usage) int64 {
	return domain.EstimateCost(s.models[modelID], usage.In, usage.Out)
}
