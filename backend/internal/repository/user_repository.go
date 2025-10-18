package repository

import (
	"context"

	"github.com/keu-5/muzee/backend/ent"
	"github.com/keu-5/muzee/backend/ent/user"
	"github.com/keu-5/muzee/backend/internal/domain"
)

type GetUserInput struct {
	ID int64
}

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}

type userRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) UserRepository {
	return &userRepository{client: client}
}

func (r *userRepository) Create(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	user, err := r.client.User.Create().
		SetEmail(email).
		SetPasswordHash(passwordHash).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return &domain.User{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := r.client.User.
		Query().
		Where(user.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	u, err := r.client.User.
		Query().
		Where(user.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}, nil
}
