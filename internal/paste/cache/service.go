package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

var ErrCacheMiss = errors.New("cache miss")

type Service struct {
	cache    *redis.Pool
	cacheTTL time.Duration
}

func NewService(cache *redis.Pool, cacheTTL time.Duration) *Service {
	return &Service{
		cache:    cache,
		cacheTTL: cacheTTL,
	}
}

func pasteKey(id string) string {
	return "paste:" + id
}

func (s *Service) GetPaste(ctx context.Context, id string) (string, error) {
	if s.cache == nil {
		return "", errors.New("cache not initialized")
	}

	conn, err := s.cache.GetContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get connection from cache: %w", err)
	}
	defer conn.Close()

	key := pasteKey(id)
	val, err := redis.String(redis.DoContext(conn, ctx, "GET", key))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
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
	if s.cache == nil {
		return errors.New("cache not initialized")
	}

	conn, err := s.cache.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection from cache: %w", err)
	}
	defer conn.Close()

	key := pasteKey(id)
	ttlSec := int64(s.cacheTTL.Seconds())
	if ttlSec > 0 {
		_, err = redis.DoContext(conn, ctx, "SET", key, text, "EX", ttlSec)
	} else {
		_, err = redis.DoContext(conn, ctx, "SET", key, text)
	}
	if err != nil {
		return fmt.Errorf("failed to set paste: %w", err)
	}

	return nil
}
