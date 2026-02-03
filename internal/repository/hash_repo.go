package repository

import (
	"context"

	"github.com/uptrace/bun"

	"task1/internal/model"
)

type hashRepo struct {
	db *bun.DB
}

func NewHashRepository(db *bun.DB) HashRepository {
	return &hashRepo{db: db}
}

func (r *hashRepo) Create(ctx context.Context, hash *model.Hash) error {
	_, err := r.db.NewInsert().Model(hash).Exec(ctx)
	return err
}

func (r *hashRepo) CreateBatch(ctx context.Context, hashes []*model.Hash) error {
	_, err := r.db.NewInsert().Model(&hashes).On("CONFLICT (md5_hash) DO NOTHING").Exec(ctx)
	return err
}

func (r *hashRepo) GetByMD5(ctx context.Context, md5 string) (*model.Hash, error) {
	var hash model.Hash
	err := r.db.NewSelect().Model(&hash).Where("md5_hash = ?", md5).Scan(ctx)
	return &hash, err
}

func (r *hashRepo) List(ctx context.Context, page, size int, sourceFile string) ([]model.Hash, int, error) {
	var hashes []model.Hash
	query := r.db.NewSelect().Model(&hashes).
		Order("id DESC").
		Offset((page - 1) * size).
		Limit(size)

	if sourceFile != "" {
		query.Where("source_file = ?", sourceFile)
	}

	count, err := query.ScanAndCount(ctx)
	return hashes, count, err
}

func (r *hashRepo) Count(ctx context.Context) (int64, error) {
	count, err := r.db.NewSelect().Model((*model.Hash)(nil)).Count(ctx)
	return int64(count), err
}
