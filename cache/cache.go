package cache

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/garyburd/redigo/redis"
	"github.com/sprungknoedl/reputile/lib"
	"golang.org/x/net/context"
)

type CacheFunc func(ctx context.Context, key string) (string, error)

func NewCache(url string) (*redis.Pool, error) {
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 120 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(url)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	conn := pool.Get()
	_, err := conn.Do("PING")
	conn.Close()
	return pool, err
}

func String(ctx context.Context, key string, calc CacheFunc) (string, error) {
	pool := ctx.Value(lib.CacheKey).(*redis.Pool)
	conn := pool.Get()
	defer conn.Close()

	value, err := redis.String(conn.Do("GET", key))
	if err == nil {
		// cache hit
		logrus.Printf("cache hit for %q", key)
		return value, nil
	}

	if err != nil && err != redis.ErrNil {
		// read redis error happened
		return "", err
	}

	// cache miss, so we have to calculate the value with the provided function
	logrus.Printf("cache miss for %q", key)
	value, err = calc(ctx, key)
	if err != nil {
		// error during calculate, don't store that in cache ;)
		return "", err
	}

	_, err = conn.Do("SET", key, value)
	conn.Do("EXPIRE", key, 3600)
	return value, err
}

func GetInt(ctx context.Context, key string) int {
	pool := ctx.Value(lib.CacheKey).(*redis.Pool)
	conn := pool.Get()
	defer conn.Close()

	value, err := redis.Int(conn.Do("GET", key))
	if err != nil {
		return 0
	}
	return value
}

func SetInt(ctx context.Context, key string, value int) {
	pool := ctx.Value(lib.CacheKey).(*redis.Pool)
	conn := pool.Get()
	defer conn.Close()

	conn.Do("SET", key, value)
}

func Incr(ctx context.Context, key string) {
	pool := ctx.Value(lib.CacheKey).(*redis.Pool)
	conn := pool.Get()
	defer conn.Close()

	conn.Do("INCR", key)
}
