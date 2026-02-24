package handlers

import (
	"context"

	"pierakladnia/internal/db"
)

type contextKey string

const UserContextKey contextKey = "user"
const ActiveProjectContextKey contextKey = "active_project"
const UserProjectsContextKey contextKey = "user_projects"

func GetUserFromContext(ctx context.Context) *db.User {
	if user, ok := ctx.Value(UserContextKey).(*db.User); ok {
		return user
	}
	return nil
}

func GetActiveProjectFromContext(ctx context.Context) *db.Project {
	if p, ok := ctx.Value(ActiveProjectContextKey).(*db.Project); ok {
		return p
	}
	return nil
}

func GetUserProjectsFromContext(ctx context.Context) []*db.Project {
	if p, ok := ctx.Value(UserProjectsContextKey).([]*db.Project); ok {
		return p
	}
	return nil
}
