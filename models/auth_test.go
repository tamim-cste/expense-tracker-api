package models_test

import (
	"testing"

	"expense-tracker-api/models"
)

// TestValidateUserInput tests the user registration validation logic.
func TestValidateUserInput(t *testing.T) {
	tests := []struct {
		name     string
		userName string
		email    string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid input",
			userName: "John Doe",
			email:    "john@example.com",
			password: "secret123",
			wantErr:  false,
		},
		{
			name:     "missing name",
			userName: "",
			email:    "john@example.com",
			password: "secret123",
			wantErr:  true,
			errMsg:   "Name is required",
		},
		{
			name:     "missing email",
			userName: "John",
			email:    "",
			password: "secret123",
			wantErr:  true,
			errMsg:   "Email is required",
		},
		{
			name:     "invalid email format",
			userName: "John",
			email:    "notanemail",
			password: "secret123",
			wantErr:  true,
			errMsg:   "Invalid email format",
		},
		{
			name:     "missing password",
			userName: "John",
			email:    "john@example.com",
			password: "",
			wantErr:  true,
			errMsg:   "Password is required",
		},
		{
			name:     "password too short",
			userName: "John",
			email:    "john@example.com",
			password: "abc",
			wantErr:  true,
			errMsg:   "Password must be at least 6 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := models.ValidateUserInput(tt.userName, tt.email, tt.password)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestValidateEmail tests the email format validation helper.
func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"user.name+tag@domain.org", true},
		{"notanemail", false},
		{"@nodomain.com", false},
		{"noatsign.com", false},
		{"user@nodot", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			got := models.ValidateEmail(tt.email)
			if got != tt.valid {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, got, tt.valid)
			}
		})
	}
}
