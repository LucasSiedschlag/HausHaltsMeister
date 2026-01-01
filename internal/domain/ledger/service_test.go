package ledger

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	role   string
	exists bool
}

func (f *fakeRepo) ListLedgersForUser(ctx context.Context, userID string) ([]LedgerWithRole, error) {
	return nil, nil
}

func (f *fakeRepo) CreateLedger(ctx context.Context, params CreateLedgerParams) (Ledger, error) {
	return Ledger{}, nil
}

func (f *fakeRepo) GetLedgerForUser(ctx context.Context, ledgerID, userID string) (LedgerWithRole, error) {
	return LedgerWithRole{}, ErrNotFound
}

func (f *fakeRepo) GetLedgerByID(ctx context.Context, ledgerID string) (Ledger, error) {
	return Ledger{}, ErrNotFound
}

func (f *fakeRepo) UpdateLedger(ctx context.Context, ledgerID, name string, updatedAt time.Time) (Ledger, error) {
	return Ledger{}, nil
}

func (f *fakeRepo) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return f.exists, nil
}

func (f *fakeRepo) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return f.role, nil
}

func (f *fakeRepo) ListMembers(ctx context.Context, ledgerID string) ([]Member, error) {
	return nil, nil
}

func (f *fakeRepo) AddMember(ctx context.Context, ledgerID, userID, role string, updatedAt time.Time) (Member, error) {
	return Member{}, nil
}

func (f *fakeRepo) UpdateMemberRole(ctx context.Context, ledgerID, userID, role string, updatedAt time.Time) (Member, error) {
	return Member{}, nil
}

func (f *fakeRepo) RemoveMember(ctx context.Context, ledgerID, userID string) error {
	return nil
}

func TestAddMemberRequiresOwner(t *testing.T) {
	repo := &fakeRepo{role: "viewer"}
	service := NewService(repo)

	_, err := service.AddMember(context.Background(), "user-1", "ledger-1", "user-2", "viewer")
	require.Error(t, err)
	require.Equal(t, ErrAccessDenied, err)
}

func TestGetLedgerReturnsAccessDeniedWhenExists(t *testing.T) {
	repo := &fakeRepo{exists: true}
	service := NewService(repo)

	_, err := service.GetLedger(context.Background(), "user-1", "ledger-1")
	require.Error(t, err)
	require.Equal(t, ErrAccessDenied, err)
}
