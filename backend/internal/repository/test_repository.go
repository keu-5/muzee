package repository

import (
	"context"

	"github.com/keu-5/muzee/backend/ent"
	"github.com/keu-5/muzee/backend/internal/domain"
)

type TestRepository interface {
	Create(ctx context.Context) (*domain.Test, error)
	GetAll(ctx context.Context) ([]*domain.Test, error)
}

type testRepository struct {
	client *ent.Client
}

func NewTestRepository(client *ent.Client) TestRepository {
	return &testRepository{client: client}
}

func (r *testRepository) Create(ctx context.Context) (*domain.Test, error) {
	test, err := r.client.Test.Create().Save(ctx)
	if err != nil {
		return nil, err
	}
	return &domain.Test{ID: test.ID}, nil
}

func (r *testRepository) GetAll(ctx context.Context) ([]*domain.Test, error) {
	tests, err := r.client.Test.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Test, 0, len(tests))
	for _, t := range tests {
		result = append(result, &domain.Test{ID: t.ID})
	}
	return result, nil
}
