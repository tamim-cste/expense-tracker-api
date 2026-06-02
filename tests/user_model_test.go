package tests

import (
	"os"
	"path/filepath"
	"testing"

	"expense-tracker-api/models"
)

// newTempUserCSV creates a temp CSV file with header for user tests.
// Returns the file path and a cleanup function.
func newTempUserCSV(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "users.csv")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create temp user CSV: %v", err)
	}
	f.WriteString("id,name,email,password,created_at\n")
	f.Close()
	models.SetUsersCSVPath(path)
	return path, func() {
		models.SetUsersCSVPath("")
		os.Remove(path)
	}
}

// seedUsers writes user rows directly into a CSV for test setup.
func seedUsers(t *testing.T, path string, rows string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("seedUsers: %v", err)
	}
	defer f.Close()
	f.WriteString(rows)
}

// TestGetAllUsers tests reading users from CSV.
func TestGetAllUsers(t *testing.T) {
	tests := []struct {
		name      string
		seedData  string
		wantCount int
	}{
		{
			name:      "empty CSV returns empty slice",
			seedData:  "",
			wantCount: 0,
		},
		{
			name:      "one user",
			seedData:  "1,John,john@example.com,secret123,2025-01-01T00:00:00Z\n",
			wantCount: 1,
		},
		{
			name:      "multiple users",
			seedData:  "1,John,john@example.com,pass1,2025-01-01T00:00:00Z\n2,Jane,jane@example.com,pass2,2025-01-02T00:00:00Z\n",
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := newTempUserCSV(t)
			defer cleanup()
			if tt.seedData != "" {
				seedUsers(t, path, tt.seedData)
			}

			users, err := models.GetAllUsers()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(users) != tt.wantCount {
				t.Errorf("got %d users, want %d", len(users), tt.wantCount)
			}
		})
	}
}

// TestGetUserByEmail tests finding a user by email.
func TestGetUserByEmail(t *testing.T) {
	tests := []struct {
		name      string
		seedData  string
		email     string
		wantFound bool
		wantName  string
	}{
		{
			name:      "found existing user",
			seedData:  "1,John,john@example.com,pass123,2025-01-01T00:00:00Z\n",
			email:     "john@example.com",
			wantFound: true,
			wantName:  "John",
		},
		{
			name:      "case insensitive match",
			seedData:  "1,John,john@example.com,pass123,2025-01-01T00:00:00Z\n",
			email:     "JOHN@EXAMPLE.COM",
			wantFound: true,
			wantName:  "John",
		},
		{
			name:      "user not found returns nil",
			seedData:  "1,John,john@example.com,pass123,2025-01-01T00:00:00Z\n",
			email:     "nobody@example.com",
			wantFound: false,
		},
		{
			name:      "empty CSV returns nil",
			seedData:  "",
			email:     "john@example.com",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := newTempUserCSV(t)
			defer cleanup()
			if tt.seedData != "" {
				seedUsers(t, path, tt.seedData)
			}

			user, err := models.GetUserByEmail(tt.email)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantFound {
				if user == nil {
					t.Fatal("expected user but got nil")
				}
				if user.Name != tt.wantName {
					t.Errorf("got name %q, want %q", user.Name, tt.wantName)
				}
			} else {
				if user != nil {
					t.Errorf("expected nil but got user: %v", user)
				}
			}
		})
	}
}

// TestGetUserByID tests finding a user by ID.
func TestGetUserByID(t *testing.T) {
	tests := []struct {
		name      string
		seedData  string
		id        int
		wantFound bool
		wantEmail string
	}{
		{
			name:      "found by id",
			seedData:  "1,John,john@example.com,pass123,2025-01-01T00:00:00Z\n2,Jane,jane@example.com,pass456,2025-01-02T00:00:00Z\n",
			id:        2,
			wantFound: true,
			wantEmail: "jane@example.com",
		},
		{
			name:      "id not found returns nil",
			seedData:  "1,John,john@example.com,pass123,2025-01-01T00:00:00Z\n",
			id:        99,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := newTempUserCSV(t)
			defer cleanup()
			if tt.seedData != "" {
				seedUsers(t, path, tt.seedData)
			}

			user, err := models.GetUserByID(tt.id)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantFound {
				if user == nil {
					t.Fatal("expected user but got nil")
				}
				if user.Email != tt.wantEmail {
					t.Errorf("got email %q, want %q", user.Email, tt.wantEmail)
				}
			} else {
				if user != nil {
					t.Errorf("expected nil but got user")
				}
			}
		})
	}
}

// TestGetNextUserID tests the auto-increment ID logic.
func TestGetNextUserID(t *testing.T) {
	tests := []struct {
		name     string
		seedData string
		wantID   int
	}{
		{
			name:     "empty CSV returns 1",
			seedData: "",
			wantID:   1,
		},
		{
			name:     "returns max id plus 1",
			seedData: "1,A,a@a.com,p,2025-01-01T00:00:00Z\n3,B,b@b.com,p,2025-01-01T00:00:00Z\n",
			wantID:   4,
		},
		{
			name:     "single user returns 2",
			seedData: "1,A,a@a.com,p,2025-01-01T00:00:00Z\n",
			wantID:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := newTempUserCSV(t)
			defer cleanup()
			if tt.seedData != "" {
				seedUsers(t, path, tt.seedData)
			}

			got := models.GetNextUserID()
			if got != tt.wantID {
				t.Errorf("GetNextUserID() = %d, want %d", got, tt.wantID)
			}
		})
	}
}

// TestCreateUser tests appending a new user to the CSV.
func TestCreateUser(t *testing.T) {
	tests := []struct {
		name       string
		seedData   string
		input      models.User
		wantID     int
		wantErrNil bool
	}{
		{
			name:     "create first user",
			seedData: "",
			input:    models.User{Name: "Alice", Email: "alice@example.com", Password: "pass123"},
			wantID:   1,
			wantErrNil: true,
		},
		{
			name:     "create second user gets id 2",
			seedData: "1,Bob,bob@example.com,pass,2025-01-01T00:00:00Z\n",
			input:    models.User{Name: "Alice", Email: "alice@example.com", Password: "pass123"},
			wantID:   2,
			wantErrNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := newTempUserCSV(t)
			defer cleanup()
			if tt.seedData != "" {
				seedUsers(t, path, tt.seedData)
			}

			err := models.CreateUser(&tt.input)
			if tt.wantErrNil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.input.ID != tt.wantID {
				t.Errorf("got ID %d, want %d", tt.input.ID, tt.wantID)
			}
			if tt.input.CreatedAt == "" {
				t.Error("CreatedAt should be set after create")
			}

			// Verify it was persisted
			found, _ := models.GetUserByEmail(tt.input.Email)
			if found == nil {
				t.Error("user not found in CSV after create")
			}
		})
	}
}
