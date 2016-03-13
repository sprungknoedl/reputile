package lib

import (
	"net/http"

	gorilla "github.com/gorilla/context"
	"golang.org/x/net/context"
)

type Key int

const (
	RequestKey  Key = iota
	CacheKey    Key = iota
	DatabaseKey Key = iota
	TemplateKey Key = iota
)

type wrapper struct {
	context.Context
	request *http.Request
}

func NewContext(r *http.Request) context.Context {
	return &wrapper{context.Background(), r}
}

func (ctx *wrapper) Value(key interface{}) interface{} {
	if key == RequestKey {
		return ctx.request
	}

	if val, ok := gorilla.GetOk(ctx.request, key); ok {
		return val
	}
	return ctx.Context.Value(key)
}
