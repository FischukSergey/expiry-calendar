package model

import "time"

const (
	// JWTIssuer — iss access-токена.
	JWTIssuer = "duekeep"
	// RefreshCookie — HttpOnly cookie с тем же refresh, что в JSON.
	RefreshCookie = "duekeep_refresh"
	// RefreshCookiePath ограничивает cookie ручками /api/v1/auth.
	RefreshCookiePath = "/api/v1/auth"
	// TokenTypeBearer — поле token_type в паре.
	TokenTypeBearer = "Bearer"
	// MinPasswordLen — порог register/login-валидации (пример API: secret12).
	MinPasswordLen = 8
)

// TokenPair — ответ login/register/refresh.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// RefreshSession — строка refresh_tokens. В БД только TokenHash, не сырой токен.
type RefreshSession struct {
	ID        string
	UserID    string
	FamilyID  string // одна «семья» устройств; reuse любого отозванного гасит всю family.
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	UserAgent string
	CreatedAt time.Time
}
