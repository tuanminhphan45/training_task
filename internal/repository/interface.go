package repository

import (
	"context"

	"task1/internal/model"
)

type HashRepository interface {
	Create(ctx context.Context, hash *model.Hash) error
	CreateBatch(ctx context.Context, hashes []*model.Hash) error
	GetByMD5(ctx context.Context, md5 string) (*model.Hash, error)
	List(ctx context.Context, page, size int, sourceFile string) ([]model.Hash, int, error)
	Count(ctx context.Context) (int64, error)
}
