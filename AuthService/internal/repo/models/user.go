package models

// User represents a user record stored in the database.
type User struct {
	ID           int64
	Login        string
	Email        string
	PasswordHash string
}
