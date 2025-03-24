package domain

// contextKey adalah tipe untuk key dalam context
type contextKey string

const (
	UserIDlKey       contextKey = "user_id"
	UserNameKey      contextKey = "username"
	UserEmailKey     contextKey = "email"
	AuthorizationKey contextKey = "Authorization"
)
