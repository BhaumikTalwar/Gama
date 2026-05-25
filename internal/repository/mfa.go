package repository

import (
	"context"

	db "github.com/BhaumikTalwar/Gama/internal/db/gen/sqlc"
	"github.com/google/uuid"
)

type MFARepo struct {
	*BaseRepo
}

func NewMfaRepo(base *BaseRepo) *MFARepo {
	return &MFARepo{BaseRepo: base}
}

func (r *MFARepo) GetSettings(ctx context.Context, userID uuid.UUID) (*db.GetUserMFASettingsRow, error) {
	val, err := r.db.GetUserMFASettings(ctx, userID)
	return &val, err
}

func (r *MFARepo) UpsertSettings(ctx context.Context, args db.UpsertMFASettingsParams) (*db.UserMfaSetting, error) {
	val, err := r.db.UpsertMFASettings(ctx, args)
	return &val, err
}

func (r *MFARepo) SetEnabled(ctx context.Context, args db.EnableMFAParams) error {
	err := r.db.EnableMFA(ctx, args)
	return err
}

func (r *MFARepo) Disable(ctx context.Context, userID uuid.UUID) error {
	err := r.db.DisableMFA(ctx, userID)
	return err
}
