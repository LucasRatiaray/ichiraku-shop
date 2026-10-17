package domain

import "time"

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         Role
	Status       Status
	CreatedAt    time.Time
}
