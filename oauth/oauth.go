package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
)

func generateState() (string, error) {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	state := hex.EncodeToString(buf)
	return state, nil
}

type StateCache struct {
	redis   *redis.Client
	expires time.Duration
}

func NewStateCache(redis *redis.Client, expires time.Duration) *StateCache {
	return &StateCache{
		redis:   redis,
		expires: expires,
	}
}

func (c *StateCache) SetState(ctx context.Context) (string, error) {
	state, err := generateState()
	if err != nil {
		return "", err
	}
	return state, c.redis.Set(ctx, state, "", c.expires).Err()
}

func (c *StateCache) GetDelState(ctx context.Context, state string) (string, error) {
	return c.redis.GetDel(ctx, state).Result()
}
