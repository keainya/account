package object

import (
	"time"
)

// User 用户模型
type User struct {
	ID           string    `gorm:"type:text;primaryKey" json:"id"`
	Username     string    `gorm:"type:text;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"type:text;not null" json:"-"`
	Email        string    `gorm:"type:text" json:"email"`
	Role         string    `gorm:"type:text;not null;default:user" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// App 应用模型
type App struct {
	ID           string    `gorm:"type:text;primaryKey" json:"id"`
	Name         string    `gorm:"type:text;not null" json:"name"`
	Description  string    `gorm:"type:text" json:"description"`
	ClientID     string    `gorm:"type:text;uniqueIndex;not null" json:"client_id"`
	ClientSecret string    `gorm:"type:text;not null" json:"client_secret,omitempty"`
	RedirectURIs string    `gorm:"type:text" json:"redirect_uris"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserMetadata 用户元数据模型
type UserMetadata struct {
	ID        string    `gorm:"type:text;primaryKey" json:"id"`
	UserID    string    `gorm:"type:text;not null;uniqueIndex:idx_user_app" json:"user_id"`
	AppID     string    `gorm:"type:text;not null;uniqueIndex:idx_user_app" json:"app_id"`
	Metadata  string    `gorm:"type:text" json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OAuthCode 授权码模型
type OAuthCode struct {
	Code        string    `gorm:"type:text;primaryKey" json:"code"`
	UserID      string    `gorm:"type:text;not null" json:"user_id"`
	ClientID    string    `gorm:"type:text;not null" json:"client_id"`
	RedirectURI string    `gorm:"type:text" json:"redirect_uri"`
	Scope       string    `gorm:"type:text" json:"scope"`
	ExpiresAt   time.Time `json:"expires_at"`
	Used        bool      `gorm:"not null;default:false" json:"used"`
	CreatedAt   time.Time `json:"created_at"`
}

// RefreshToken 刷新令牌模型
type RefreshToken struct {
	Token     string    `gorm:"type:text;primaryKey" json:"token"`
	UserID    string    `gorm:"type:text;not null" json:"user_id"`
	ClientID  string    `gorm:"type:text;not null" json:"client_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `gorm:"not null;default:false" json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
}
