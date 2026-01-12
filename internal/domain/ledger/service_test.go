package ledger

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	role        string
	exists      bool
	added       Member
	ledger      Ledger
	userByEmail map[string]string
}

func (f *fakeRepo) ListLedgersForUser(ctx context.Context, userID string) ([]LedgerWithRole, error) {
	return nil, nil
}

func (f *fakeRepo) CreateLedger(ctx context.Context, params CreateLedgerParams) (Ledger, error) {
	return Ledger{ID: "ledger-1", OwnerUserID: params.OwnerUserID, Name: params.Name, CurrencyCode: params.CurrencyCode}, nil
}

func (f *fakeRepo) GetLedgerForUser(ctx context.Context, ledgerID, userID string) (LedgerWithRole, error) {
	return LedgerWithRole{}, ErrNotFound
}

func (f *fakeRepo) GetLedgerByID(ctx context.Context, ledgerID string) (Ledger, error) {
	if f.ledger.ID == "" {
		return Ledger{}, ErrNotFound
	}
	return f.ledger, nil
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
	f.added = Member{LedgerID: ledgerID, UserID: userID, Role: role}
	return f.added, nil
}

func (f *fakeRepo) UpdateMemberRole(ctx context.Context, ledgerID, userID, role string, updatedAt time.Time) (Member, error) {
	return Member{LedgerID: ledgerID, UserID: userID, Role: role}, nil
}

func (f *fakeRepo) RemoveMember(ctx context.Context, ledgerID, userID string) error {
	return nil
}

func (f *fakeRepo) GetUserIDByEmail(ctx context.Context, email string) (string, error) {
	if f.userByEmail != nil {
		if userID, ok := f.userByEmail[email]; ok {
			return userID, nil
		}
	}
	return "", ErrUserNotFound
}

func TestAddMemberRequiresOwner(t *testing.T) {
	repo := &fakeRepo{role: "viewer"}
	service := NewService(repo)

	_, err := service.AddMember(context.Background(), "user-1", "ledger-1", "user-2", "viewer")
	require.Error(t, err)
	require.Equal(t, ErrAccessDenied, err)
}

func TestUpdateLedgerRequiresOwner(t *testing.T) {
	repo := &fakeRepo{role: "viewer"}
	service := NewService(repo)

	_, err := service.UpdateLedger(context.Background(), "user-1", "ledger-1", "Novo nome")
	require.Equal(t, ErrAccessDenied, err)
}

func TestDeleteLedgerRequiresOwner(t *testing.T) {
	repo := &fakeRepo{role: "viewer"}
	service := NewService(repo)

	err := service.DeleteLedger(context.Background(), "user-1", "ledger-1")
	require.Equal(t, ErrAccessDenied, err)
}

func TestListMembersAllowsViewer(t *testing.T) {
	repo := &fakeRepo{role: "viewer"}
	service := NewService(repo)

	_, err := service.ListMembers(context.Background(), "user-1", "ledger-1")
	require.NoError(t, err)
}

func TestUpdateMemberRequiresOwner(t *testing.T) {
	repo := &fakeRepo{role: "viewer"}
	service := NewService(repo)

	_, err := service.UpdateMember(context.Background(), "user-1", "ledger-1", "user-2", "editor")
	require.Equal(t, ErrAccessDenied, err)
}

func TestRemoveMemberRequiresOwner(t *testing.T) {
	repo := &fakeRepo{role: "viewer"}
	service := NewService(repo)

	err := service.RemoveMember(context.Background(), "user-1", "ledger-1", "user-2")
	require.Equal(t, ErrAccessDenied, err)
}

func TestGetLedgerReturnsAccessDeniedWhenExists(t *testing.T) {
	repo := &fakeRepo{exists: true}
	service := NewService(repo)

	_, err := service.GetLedger(context.Background(), "user-1", "ledger-1")
	require.Error(t, err)
	require.Equal(t, ErrAccessDenied, err)
}

func TestCreateLedgerSuccess(t *testing.T) {
	repo := &fakeRepo{role: "owner"}
	service := NewService(repo)

	ledger, err := service.CreateLedger(context.Background(), "user-1", "Pessoal", "BRL", true, false)
	require.NoError(t, err)
	require.Equal(t, "Pessoal", ledger.Name)
	require.Equal(t, "BRL", ledger.CurrencyCode)
}

func TestAddMemberSuccess(t *testing.T) {
	repo := &fakeRepo{role: "owner"}
	service := NewService(repo)

	member, err := service.AddMember(context.Background(), "user-1", "ledger-1", "user-2", "viewer")
	require.NoError(t, err)
	require.Equal(t, "user-2", member.UserID)
	require.Equal(t, "viewer", member.Role)
}

func TestAddMemberByEmailInvalidEmail(t *testing.T) {
	repo := &fakeRepo{role: "owner"}
	service := NewService(repo)

	_, err := service.AddMemberByEmail(context.Background(), "user-1", "ledger-1", "invalid", "viewer")
	require.Error(t, err)
	ledgerErr, ok := err.(*Error)
	require.True(t, ok)
	require.Equal(t, "VALIDATION_ERROR", ledgerErr.Code())
	require.Equal(t, "invalid", ledgerErr.Details()["email"])
}

func TestAddMemberByEmailNotFound(t *testing.T) {
	repo := &fakeRepo{role: "owner", userByEmail: map[string]string{}}
	service := NewService(repo)

	_, err := service.AddMemberByEmail(context.Background(), "user-1", "ledger-1", "user@example.com", "viewer")
	require.Error(t, err)
	ledgerErr, ok := err.(*Error)
	require.True(t, ok)
	require.Equal(t, "VALIDATION_ERROR", ledgerErr.Code())
	require.Equal(t, "not_found", ledgerErr.Details()["email"])
}

func TestAddMemberByEmailSuccess(t *testing.T) {
	repo := &fakeRepo{role: "owner", userByEmail: map[string]string{"user@example.com": "user-2"}}
	service := NewService(repo)

	member, err := service.AddMemberByEmail(context.Background(), "user-1", "ledger-1", "user@example.com", "viewer")
	require.NoError(t, err)
	require.Equal(t, "user-2", member.UserID)
	require.Equal(t, "viewer", member.Role)
}

func TestUpdateMemberRole(t *testing.T) {
	repo := &fakeRepo{role: "owner", ledger: Ledger{ID: "ledger-1", OwnerUserID: "user-1"}}
	service := NewService(repo)

	member, err := service.UpdateMember(context.Background(), "user-1", "ledger-1", "user-2", "editor")
	require.NoError(t, err)
	require.Equal(t, "editor", member.Role)
}

func TestRemoveMember(t *testing.T) {
	repo := &fakeRepo{role: "owner", ledger: Ledger{ID: "ledger-1", OwnerUserID: "user-1"}}
	service := NewService(repo)

	err := service.RemoveMember(context.Background(), "user-1", "ledger-1", "user-2")
	require.NoError(t, err)
}
