package audit

import (
	"context"
	"time"
)

type Event struct {
	LedgerID  string
	UserID    string
	Action    string
	EntityID  *string
	IP        string
	UserAgent string
	CreatedAt time.Time
}

type Recorder interface {
	Record(ctx context.Context, event Event) error
}

type NopRecorder struct{}

func (NopRecorder) Record(ctx context.Context, event Event) error {
	return nil
}
