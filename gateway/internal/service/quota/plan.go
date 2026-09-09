package quota

import (
	"time"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"
)

type PlanCode string

const (
	PlanFree PlanCode = "free"
	PlanPlus PlanCode = "plus"
)

type Plan struct {
	Code              PlanCode
	DailyMessages     int
	MonthlyCostMicros int64
	AllowedTiers      []provider.Tier
	ConcurrentStreams int
}

var plans = map[PlanCode]Plan{
	PlanFree: {
		Code:              PlanFree,
		DailyMessages:     30,
		MonthlyCostMicros: 0,
		AllowedTiers:      []provider.Tier{provider.TierLocal},
		ConcurrentStreams: 1,
	},
	PlanPlus: {
		Code:              PlanPlus,
		DailyMessages:     300,
		MonthlyCostMicros: 5_000_000,
		AllowedTiers:      []provider.Tier{provider.TierLocal, provider.TierFrontier},
		ConcurrentStreams: 3,
	},
}

func PlanFor(code PlanCode) Plan {
	if p, ok := plans[code]; ok {
		return p
	}
	return plans[PlanFree]
}

func (p Plan) Allows(tier provider.Tier) bool {
	for _, allowed := range p.AllowedTiers {
		if allowed == tier {
			return true
		}
	}
	return false
}

// Subscription expiry is evaluated on read, so nothing breaks when a job fails
// to run: an expired row simply means the free plan.
type Subscription struct {
	PlanCode  PlanCode
	ExpiresAt *time.Time
}

func (s Subscription) Active(now time.Time) bool {
	return s.ExpiresAt == nil || s.ExpiresAt.After(now)
}
