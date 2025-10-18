package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SignupSessionData struct {
	PasswordHash string `json:"password_hash"`
	Code         string `json:"code"`
	CreatedAt    int64  `json:"created_at"`
}

type RefreshTokenData struct {
	UserID    int64 `json:"user_id"`
	CreatedAt int64 `json:"created_at"`
}

type SessionHelper struct {
	redisClient *redis.Client
}

func NewSessionHelper(redisClient *redis.Client) *SessionHelper {
	return &SessionHelper{
		redisClient: redisClient,
	}
}

// CheckRateLimit checks if the email has exceeded the rate limit for sending codes
func (s *SessionHelper) CheckRateLimit(ctx context.Context, email string) error {
	rateLimitKey := fmt.Sprintf("rate_limit:send_code:%s", email)
	count, err := s.redisClient.Incr(ctx, rateLimitKey).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		s.redisClient.Expire(ctx, rateLimitKey, 5*time.Minute)
	}
	if count > 3 {
		return fmt.Errorf("rate limit exceeded")
	}
	return nil
}

// CheckLoginRateLimit checks if the email has exceeded the rate limit for login attempts
func (s *SessionHelper) CheckLoginRateLimit(ctx context.Context, email string) error {
	rateLimitKey := fmt.Sprintf("rate_limit:login:%s", email)
	count, err := s.redisClient.Incr(ctx, rateLimitKey).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		s.redisClient.Expire(ctx, rateLimitKey, 15*time.Minute)
	}
	if count > 5 {
		return fmt.Errorf("rate limit exceeded")
	}
	return nil
}

// SaveSignupSession saves the signup session data to Redis
func (s *SessionHelper) SaveSignupSession(ctx context.Context, email, passwordHash, code string) error {
	sessionData := SignupSessionData{
		PasswordHash: passwordHash,
		Code:         code,
		CreatedAt:    time.Now().Unix(),
	}
	dataJSON, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("signup:%s", email)
	return s.redisClient.Set(ctx, key, dataJSON, 15*time.Minute).Err()
}

// GetSignupSession retrieves the signup session data from Redis
func (s *SessionHelper) GetSignupSession(ctx context.Context, email string) (*SignupSessionData, error) {
	key := fmt.Sprintf("signup:%s", email)
	data, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var sessionData SignupSessionData
	if err := json.Unmarshal([]byte(data), &sessionData); err != nil {
		return nil, err
	}

	return &sessionData, nil
}

// DeleteSignupSession deletes the signup session data from Redis
func (s *SessionHelper) DeleteSignupSession(ctx context.Context, email string) error {
	key := fmt.Sprintf("signup:%s", email)
	return s.redisClient.Del(ctx, key).Err()
}

// SaveRefreshToken saves the refresh token data to Redis
func (s *SessionHelper) SaveRefreshToken(ctx context.Context, token string, userID int64) error {
	tokenData := RefreshTokenData{
		UserID:    userID,
		CreatedAt: time.Now().Unix(),
	}
	dataJSON, err := json.Marshal(tokenData)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("refresh_token:%s", token)
	return s.redisClient.Set(ctx, key, dataJSON, 30*24*time.Hour).Err()
}

// GetRefreshToken retrieves the refresh token data from Redis
func (s *SessionHelper) GetRefreshToken(ctx context.Context, token string) (*RefreshTokenData, error) {
	key := fmt.Sprintf("refresh_token:%s", token)
	data, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var tokenData RefreshTokenData
	if err := json.Unmarshal([]byte(data), &tokenData); err != nil {
		return nil, err
	}

	return &tokenData, nil
}

// DeleteRefreshToken deletes the refresh token data from Redis
func (s *SessionHelper) DeleteRefreshToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return s.redisClient.Del(ctx, key).Err()
}
