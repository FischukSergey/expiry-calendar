package model

import "time"

// Role — роль пользователя. Register создаёт admin своих данных.
type Role string

const (
	// RoleAdmin — CRUD своих записей и категорий. Справочник типов не меняет.
	RoleAdmin Role = "admin"
	// RoleViewer — только чтение своих данных.
	RoleViewer Role = "viewer"
	// RoleAdministrator — как admin, плюс мутации общего item_kinds. Регистрация её не выдаёт.
	RoleAdministrator Role = "administrator"
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
