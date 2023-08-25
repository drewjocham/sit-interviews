package auth

import (
	"context"
	"github.com/pkg/errors"
	log "interviews/pkg/logger"
	"sync"
	"time"
)

var (
	ErrorTokenCacheNotAvailable = errors.New("token cache not available")
	ErrCacheTooOld              = errors.New("token cache too old")
	ErrRefreshCache             = errors.New("error refreshing access token cache")
)

type TokenCache struct {
	tokenRepo TokenRepo
	updatedAt time.Time
	cache     map[string]string
	cacheLock sync.RWMutex
}

type TokenRepo interface {
	GetAllTokens(ctx context.Context) (map[string]string, error)
}

func (c *TokenCache) AddToken(ctx context.Context, email, token string) error {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	c.cache[email] = token

	return nil
}

func (c *TokenCache) RemoveToken(ctx context.Context, email string) error {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	_, ok := c.cache[email]
	if !ok {

		return ErrInvalidToken
	}
	delete(c.cache, email)

	return nil
}

func (c *TokenCache) TokenExists(email, token string) bool {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	cachedToken, ok := c.cache[email]
	if !ok {
		return false
	}

	return token == cachedToken
}

func (c *TokenCache) Init(ctx context.Context, maxRetries int, retryPeriod time.Duration) error {

	var retries int
	// try request ad templates until maxRetries
	for retries = 0; retries < maxRetries; retries++ {
		err := c.load(ctx)
		if err == nil {
			break
		}

		if err != nil {
			time.Sleep(time.Duration(retries+1) * retryPeriod)

			continue
		}

		return err
	}

	if retries == maxRetries {
		return ErrorTokenCacheNotAvailable
	}

	return nil

}

func (c *TokenCache) load(ctx context.Context) error {
	clog := log.GetLoggerFromContext(ctx)

	cacheMap := make(map[string]string, 1)

	cache, err := c.tokenRepo.GetAllTokens(ctx)
	if err != nil {
		clog.ErrorCtx(err, log.Ctx{
			"op": "unable to get tokens from the database",
		})

		return err
	}

	for k, v := range cache {
		cacheMap[k] = v
	}

	// swap
	c.cacheLock.Lock()
	c.cache = cacheMap
	c.updatedAt = time.Now()
	c.cacheLock.Unlock()

	return nil
}

func (c *TokenCache) StartRefresh(ctx context.Context, refreshPeriod, validityThreshold time.Duration) error {
	ticker := time.NewTicker(refreshPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := c.load(ctx)
			if err != nil {
				updatedAt := c.getUpdatedAt()

				//TODO: create metric here
				log.ErrorCtx(ErrRefreshCache, log.Ctx{"error": err.Error(), "updatedAt": updatedAt})

				if time.Now().After(updatedAt.Add(validityThreshold)) {
					return ErrCacheTooOld
				}
			}
		}
	}
}

func (c *TokenCache) getUpdatedAt() time.Time {
	c.cacheLock.RLock()
	defer c.cacheLock.RUnlock()

	return c.updatedAt
}

func NewTokenCache(tokenRepo TokenRepo) *TokenCache {
	return &TokenCache{
		cache:     make(map[string]string, 1),
		cacheLock: sync.RWMutex{},
		updatedAt: time.Time{},
		tokenRepo: tokenRepo,
	}
}
