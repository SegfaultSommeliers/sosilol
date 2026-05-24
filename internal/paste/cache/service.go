package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache miss")

type Service struct {
	client   *redis.Client
	cacheTTL time.Duration
}

func NewService(client *redis.Client, cacheTTL time.Duration) *Service {
	return &Service{
		client:   client,
		cacheTTL: cacheTTL,
	}
}

func pasteKey(id string) string {
	return "paste:" + id
}

func (s *Service) GetPaste(ctx context.Context, id string) (string, error) {
	if s.client == nil {
		return "", errors.New("cache not initialized")
	}

	key := pasteKey(id)
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrCacheMiss
		}

		return "", fmt.Errorf("failed to get paste from cache: %w", err)
	}

	return val, nil
}

func (s *Service) SetPaste(
	ctx context.Context,
	id string,
	text string,
) error {
	if s.client == nil {
		return errors.New("cache not initialized")
	}

	key := pasteKey(id)
	if err := s.client.Set(ctx, key, text, s.cacheTTL).Err(); err != nil {
		return fmt.Errorf("failed to set paste: %w", err)
	}

	return nil
}
