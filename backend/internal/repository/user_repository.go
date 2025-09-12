package repository

import (
	"context"
	"time"

	"github.com/keu-5/muzee/backend/ent"
	"github.com/keu-5/muzee/backend/ent/user"
)

type UserRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) *UserRepository {
	return &UserRepository{
		client: client,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, username, email, password string) (*ent.User, error) {
	return r.client.User.Create().
		SetUsername(username).
		SetEmail(email).
		SetPassword(password).
		Save(ctx)
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.ID(id), user.DeletedAtIsNil()).
		Only(ctx)
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.Username(username), user.DeletedAtIsNil()).
		Only(ctx)
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.Email(email), user.DeletedAtIsNil()).
		Only(ctx)
}

func (r *UserRepository) UpdateUser(ctx context.Context, id int, username, email, password string) (*ent.User, error) {
	return r.client.User.UpdateOneID(id).
		SetUsername(username).
		SetEmail(email).
		SetPassword(password).
		Save(ctx)
}

func (r *UserRepository) DeleteUser(ctx context.Context, id int) error {
	return r.client.User.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
