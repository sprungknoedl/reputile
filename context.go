package main

import (
	"net/http"

	gorilla "github.com/gorilla/context"
	"golang.org/x/net/context"
)

type key int

const (
	requestKey     key = iota
	cacheKey       key = iota
	databaseKey    key = iota
	transactionKey key = iota
)

type wrapper struct {
	context.Context
	request *http.Request
}

func NewContext(r *http.Request) context.Context {
	return &wrapper{context.Background(), r}
}

func (ctx *wrapper) Value(key interface{}) interface{} {
	if key == requestKey {
		return ctx.request
	}

	if val, ok := gorilla.GetOk(ctx.request, key); ok {
		return val
	}
	return ctx.Context.Value(key)
}
