package core

import "context"

// contextKey is a type-safe key for context values
type contextKey string

const (
	// WorkspaceIDKey is the context key for workspace ID
	WorkspaceIDKey contextKey = "workspace_id"
)

// ContextWithWorkspaceID returns a context with the workspace ID attached
func ContextWithWorkspaceID(ctx context.Context, workspaceID string) context.Context {
	return context.WithValue(ctx, WorkspaceIDKey, workspaceID)
}

// WorkspaceIDFromContext extracts the workspace ID from a context
func WorkspaceIDFromContext(ctx context.Context) string {
	if v := ctx.Value(WorkspaceIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return "default"
}
