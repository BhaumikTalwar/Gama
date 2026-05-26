package repository

import (
	"context"
	"fmt"
	"time"

	db "github.com/BhaumikTalwar/Gama/internal/db/gen/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepo struct {
	*BaseRepo
}

func NewUserRepo(base *BaseRepo) *UserRepo {
	return &UserRepo{BaseRepo: base}
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*db.User, error) {
	user, err := r.db.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*db.User, error) {
	user, err := r.db.GetUserByEmail(ctx, email)
	if err == pgx.ErrNoRows {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user ID by email: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) GetByPhone(ctx context.Context, phoneNumber string) (*db.User, error) {
	user, err := r.db.GetUserByPhone(ctx, phoneNumber)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*db.User, error) {
	user, err := r.db.GetUserByUsername(ctx, username)
	if err == pgx.ErrNoRows {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user ID by email: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) Create(ctx context.Context, arg db.CreateUserParams) (*db.User, error) {
	user, err := r.db.CreateUser(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	_ = r.TouchEntity(ctx, "user")
	_ = r.InvalidateCache(ctx, r.keyGen.Simple("user", "global", "list"))

	return &user, nil
}

func (r *UserRepo) Update(ctx context.Context, arg db.UpdateUserParams) (*db.User, error) {
	_, err := r.db.GetUserByID(ctx, arg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for update: %w", err)
	}

	updatedUser, err := r.db.UpdateUser(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	_ = r.TouchEntity(ctx, "user")
	_ = r.InvalidateCache(ctx, r.keyGen.Simple("user", "global", "list"))

	return &updatedUser, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, arg db.UpdateUserPasswordParams) error {
	err := r.db.UpdateUserPassword(ctx, arg)
	if err != nil {
		return err
	}

	_ = r.TouchEntity(ctx, "user")
	_ = r.InvalidateCache(ctx, r.keyGen.Simple("user", "global", "list"))

	return nil
}

func (r *UserRepo) ListUsers(ctx context.Context, arg db.ListUsersParams) ([]db.User, error) {
	return r.db.ListUsers(ctx, arg)
}

func (r *UserRepo) CountUsers(ctx context.Context) (int64, error) {
	return r.db.CountUsers(ctx)
}

func (r *UserRepo) SearchUsers(ctx context.Context, arg db.SearchUsersParams) ([]db.User, error) {
	versions, err := r.GetVersionsForEntities(ctx, []string{"user", "role"})
	if err != nil {
		versions = nil
	}

	key := r.keyGen.WithVersions(versions, r.keyGen.WithParamMap("search", map[string]any{
		"q":      arg.SearchQuery,
		"limit":  arg.Limitval,
		"offset": arg.Offsetval,
	}))

	return Fetch(ctx, r.BaseRepo, key, 5*time.Minute, func() ([]db.User, error) {
		return r.db.SearchUsers(ctx, arg)
	})
}

func (r *UserRepo) CountUsersBySearch(ctx context.Context, searchQuery *string) (int64, error) {
	versions, err := r.GetVersionsForEntities(ctx, []string{"user"})
	if err != nil {
		versions = nil
	}

	key := r.keyGen.WithVersions(versions, r.keyGen.WithParamMap("count", map[string]any{
		"q": searchQuery,
	}))

	result, err := Fetch(ctx, r.BaseRepo, key, 5*time.Minute, func() (int64, error) {
		return r.db.CountUsersBySearch(ctx, searchQuery)
	})
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *UserRepo) SoftDeleteUser(ctx context.Context, id uuid.UUID) error {
	user, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("could not get user to delete: %w", err)
	}

	if err := r.db.SoftDeleteUser(ctx, user.ID); err != nil {
		return err
	}

	_ = r.TouchEntity(ctx, "user")
	_ = r.InvalidateCache(ctx, r.keyGen.Simple("user", "global", "list"))

	return nil
}

func (r *UserRepo) GetUsersByIDs(ctx context.Context, userIDs []uuid.UUID) ([]*db.User, error) {
	users := make([]*db.User, 0, len(userIDs))
	for _, id := range userIDs {
		user, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get user with id %s: %w", id, err)
		}

		users = append(users, user)
	}

	return users, nil
}
