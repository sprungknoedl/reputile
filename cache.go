package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/garyburd/redigo/redis"
	"golang.org/x/net/context"
)

type CacheFunc func(ctx context.Context, key string) ([]byte, error)

func Cache(ctx context.Context, key string, calc CacheFunc) ([]byte, error) {
	conn := ctx.Value(cacheKey).(redis.Conn)
	bytes, err := redis.Bytes(conn.Do("GET", key))
	if err == nil {
		// cache hit
		logrus.Printf("cache hit for %q", key)
		return bytes, nil
	}

	if err != nil && err != redis.ErrNil {
		// read redis error happened
		return []byte{}, err
	}

	// cache miss, so we have to calculate the value with the provided function
	logrus.Printf("cache miss for %q", key)
	bytes, err = calc(ctx, key)
	if err != nil {
		// error during calculate, don't store that in cache ;)
		return []byte{}, err
	}

	_, err = conn.Do("SET", key, bytes)
	conn.Do("EXPIRE", key, 3600)
	return bytes, err
}
