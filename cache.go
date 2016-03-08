package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/garyburd/redigo/redis"
	"golang.org/x/net/context"
)

type CacheFunc func(ctx context.Context, key string) (string, error)

func Cache(ctx context.Context, key string, calc CacheFunc) (string, error) {
	conn := ctx.Value(cacheKey).(redis.Conn)
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

func GetCounter(ctx context.Context, key string) int {
	conn := ctx.Value(cacheKey).(redis.Conn)
	value, err := redis.Int(conn.Do("GET", key))
	if err != nil {
		logrus.Errorf("counter get %q: %v", key, err)
		return 0
	}
	return value
}

func IncrCounter(ctx context.Context, key string) {
	conn := ctx.Value(cacheKey).(redis.Conn)
	_, err := conn.Do("INCR", key)
	if err != nil {
		logrus.Errorf("counter incr %q: %v", key, err)
	}
}
