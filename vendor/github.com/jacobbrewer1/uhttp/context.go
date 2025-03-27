package uhttp

// ContextKey is key type for contexts.
type ContextKey string

const (
	// authHeader is the Authorization header in a HTTP request.
	authHeader = "Authorization"
)

var (
	// authHeaderKey is the context key to the value of the Authorization HTTP request.
	authHeaderKey = ContextKey(authHeader)
)
