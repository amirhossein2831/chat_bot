package domain

import "github.com/amirhossein2831/chat_bot/gateway/internal/config"

// EstimateCost returns USD micro-cents. Integer arithmetic throughout: money
// never touches a float. Rounds up so a partial thousand is never free.
func EstimateCost(m config.ModelConfig, in, out int) int64 {
	return ceilDiv(int64(in)*m.CostPer1kInUSDMicros, 1000) +
		ceilDiv(int64(out)*m.CostPer1kOutUSDMicros, 1000)
}

func ceilDiv(numerator, denominator int64) int64 {
	if numerator <= 0 {
		return 0
	}
	return (numerator + denominator - 1) / denominator
}
