package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/keu-5/muzee/backend/internal/database"
	"github.com/keu-5/muzee/backend/internal/db"
)

type UserRepository struct {
	queries *db.Queries
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		queries: database.GetQueries(),
	}
}

func (r *UserRepository) CreateUser(params db.CreateUserParams) (db.User, error) {
	return r.queries.CreateUser(context.Background(), params)
}

func (r *UserRepository) GetUserByID(id int32) (db.User, error) {
	user, err := r.queries.GetUserByID(context.Background(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, nil
		}
		return db.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetUserByUsername(username string) (db.User, error) {
	user, err := r.queries.GetUserByUsername(context.Background(), username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, nil
		}
		return db.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetUserByEmail(email string) (db.User, error) {
	user, err := r.queries.GetUserByEmail(context.Background(), email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, nil
		}
		return db.User{}, err
	}
	return user, nil
}
