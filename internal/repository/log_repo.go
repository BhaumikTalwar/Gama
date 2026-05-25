package repository

import (
	"context"

	"github.com/BhaumikTalwar/Gama/internal/db/gen/sqlc"
)

type LogRepo struct {
	*BaseRepo
}

func NewLogRepo(base *BaseRepo) *LogRepo {
	return &LogRepo{BaseRepo: base}
}

func (r *LogRepo) CreateUserLog(ctx context.Context, arg db.CreateUserLogParams) error {
	return r.db.CreateUserLog(ctx, arg)
}

func (r *LogRepo) GetUserLogs(ctx context.Context, arg db.GetUserLogsParams) ([]db.UserLog, error) {
	return r.db.GetUserLogs(ctx, arg)
}
