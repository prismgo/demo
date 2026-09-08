package models

import "time"

// User is the starter application's default user model.
//
// The model is intentionally small so new applications can choose their own
// authentication and persistence strategy without removing generated fields.
type User struct {
	// ID is the primary identifier for the user record.
	ID uint `json:"id"`

	// Name is the display name shown in application interfaces.
	Name string `json:"name"`

	// Email is the unique email address used by the application.
	Email string `json:"email"`

	// EmailVerifiedAt records when the email address was verified.
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`

	// Password stores the hashed password for password-based authentication.
	Password string `json:"-"`

	// RememberToken stores the persistent login token used by remember-me sessions.
	RememberToken string `json:"remember_token,omitempty"`

	// CreatedAt records when the user was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt records when the user was last changed.
	UpdatedAt time.Time `json:"updated_at"`

	// DeletedAt records soft deletion without removing the row immediately.
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
