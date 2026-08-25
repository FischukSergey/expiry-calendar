package model

import "time"

// Role — глобальная роль v1 (не per-org; org_id появится в Sprint 7).
type Role string

const (
	// RoleAdmin — полный CRUD справочников и записей.
	RoleAdmin Role = "admin"
	// RoleViewer — только чтение.
	RoleViewer Role = "viewer"
)

// User — строка users. PasswordHash не отдаём в JSON.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Role         Role      `json:"role"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"-"`
}

// PublicUser — ответ GET /me.
type PublicUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  Role   `json:"role"`
}

// Public возвращает поля контракта /me без хеша.
func (u User) Public() PublicUser {
	return PublicUser{ID: u.ID, Email: u.Email, Role: u.Role}
}
