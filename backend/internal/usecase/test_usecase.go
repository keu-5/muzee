package usecase

import (
	"context"

	"github.com/keu-5/muzee/backend/internal/domain"
	"github.com/keu-5/muzee/backend/internal/repository"
)

type TestUsecase interface {
	CreateTest(ctx context.Context) (*domain.Test, error)
	GetAllTests(ctx context.Context) ([]*domain.Test, error)
}

type testUsecase struct {
	repo repository.TestRepository
}

func NewTestUsecase(repo repository.TestRepository) TestUsecase {
	return &testUsecase{repo: repo}
}

func (u *testUsecase) CreateTest(ctx context.Context) (*domain.Test, error) {
	return u.repo.Create(ctx)
}

func (u *testUsecase) GetAllTests(ctx context.Context) ([]*domain.Test, error) {
	return u.repo.GetAll(ctx)
}
