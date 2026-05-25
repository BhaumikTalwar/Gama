package repository

import (
	"context"

	db "github.com/BhaumikTalwar/Gama/internal/db/gen/sqlc"
	"github.com/google/uuid"
)

type TokenRepo struct {
	*BaseRepo
}

func NewTokenRepo(base *BaseRepo) *TokenRepo {
	return &TokenRepo{BaseRepo: base}
}

func (r *TokenRepo) CreateRefreshToken(ctx context.Context, arg db.CreateRefreshTokenParams) (*db.RefreshToken, error) {
	token, err := r.db.CreateRefreshToken(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *TokenRepo) GetRefreshToken(ctx context.Context, tokenHash string) (*db.RefreshToken, error) {
	token, err := r.db.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *TokenRepo) RevokeRefreshToken(ctx context.Context, arg db.RevokeRefreshTokenParams) error {
	return r.db.RevokeRefreshToken(ctx, arg)
}

func (r *TokenRepo) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	return r.db.RevokeAllUserTokens(ctx, userID)
}

func (r *TokenRepo) CreateVerificationToken(ctx context.Context, arg db.CreateVerificationTokenParams) (*db.VerificationToken, error) {
	token, err := r.db.CreateVerificationToken(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *TokenRepo) GetVerificationTokenByHash(ctx context.Context, arg db.GetVerificationTokenByHashParams) (*db.VerificationToken, error) {
	token, err := r.db.GetVerificationTokenByHash(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *TokenRepo) GetLatestVerificationTokenForUser(ctx context.Context, arg db.GetLatestVerificationTokenForUserParams) (*db.VerificationToken, error) {
	token, err := r.db.GetLatestVerificationTokenForUser(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *TokenRepo) GetVerificationTokenByExternalID(ctx context.Context, arg db.GetVerificationTokenByExternalIDParams) (*db.VerificationToken, error) {
	token, err := r.db.GetVerificationTokenByExternalID(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *TokenRepo) MarkTokenUsed(ctx context.Context, id uuid.UUID) error {
	return r.db.MarkTokenUsed(ctx, id)
}
