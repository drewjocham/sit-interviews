package data

import (
	"context"
	"net/http"
)

type UsersContext struct {
}
type contextKey string

const userContextKey = contextKey("user")

func (m *UsersContext) ContextSetUser(r *http.Request, user *User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (m *UsersContext) ContextGetUser(r *http.Request) *User {
	user, ok := r.Context().Value(userContextKey).(*User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
