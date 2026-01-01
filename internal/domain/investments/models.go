package investments

import "time"

type Summary struct {
	From               time.Time
	To                 time.Time
	TotalContributions int64
	TotalRedemptions   int64
	TotalEarnings      int64
	TotalLosses        int64
	NetVariation       int64
}

type FlowCategoryIDs struct {
	ContributionOut string
	ContributionIn  string
	RedemptionOut   string
	RedemptionIn    string
	EarningsIn      string
	LossesOut       *string
}
