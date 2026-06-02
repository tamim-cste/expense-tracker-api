package models

import (
	"encoding/csv"
	"path/filepath"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web"
)

// User represents a registered user in the system.
type User struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
}

// overrideUsersCSVPath allows tests to redirect CSV reads/writes to a temp file.
var overrideUsersCSVPath string

// SetUsersCSVPath overrides the CSV path used by all user model functions.
// Pass an empty string to reset to the config default. Used only in tests.
func SetUsersCSVPath(path string) {
	overrideUsersCSVPath = path
}

// getUsersCSVPath returns the active CSV file path for users.
// If SetUsersCSVPath has been called (e.g. in tests), that path takes priority.
func getUsersCSVPath() string {
	if overrideUsersCSVPath != "" {
		return overrideUsersCSVPath
	}
	path, _ := web.AppConfig.String("users_csv_path")
	if path == "" {
		path = "data/users.csv"
	}
	return path
}

// ensureUsersCSV creates the CSV file with headers if it does not already exist.
func ensureUsersCSV(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		w := csv.NewWriter(f)
		if err := w.Write([]string{"id", "name", "email", "password", "created_at"}); err != nil {
			return err
		}
		w.Flush()
		return w.Error()
	}
	return nil
}

// GetAllUsers reads and returns all users from the CSV file.
func GetAllUsers() ([]User, error) {
	path := getUsersCSVPath()
	if err := ensureUsersCSV(path); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var users []User
	for i, record := range records {
		if i == 0 {
			continue
		}
		if len(record) < 5 {
			continue
		}
		id, _ := strconv.Atoi(record[0])
		users = append(users, User{
			ID:        id,
			Name:      record[1],
			Email:     record[2],
			Password:  record[3],
			CreatedAt: record[4],
		})
	}
	return users, nil
}

// GetUserByEmail finds a user by their email address (case-insensitive).
func GetUserByEmail(email string) (*User, error) {
	users, err := GetAllUsers()
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if strings.EqualFold(u.Email, email) {
			return &u, nil
		}
	}
	return nil, nil
}

// GetUserByID finds a user by their numeric ID.
func GetUserByID(id int) (*User, error) {
	users, err := GetAllUsers()
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, nil
}

// GetNextUserID returns the next available user ID (max existing ID + 1).
func GetNextUserID() int {
	users, err := GetAllUsers()
	if err != nil || len(users) == 0 {
		return 1
	}
	max := 0
	for _, u := range users {
		if u.ID > max {
			max = u.ID
		}
	}
	return max + 1
}

// CreateUser appends a new user row to the CSV file.
func CreateUser(user *User) error {
	path := getUsersCSVPath()
	if err := ensureUsersCSV(path); err != nil {
		return err
	}

	user.ID = GetNextUserID()
	user.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	err = w.Write([]string{
		strconv.Itoa(user.ID),
		user.Name,
		user.Email,
		user.Password,
		user.CreatedAt,
	})
	if err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

// ValidateEmail checks if the email format is valid.
func ValidateEmail(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}
	if !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

// ValidateUserInput validates registration input fields.
func ValidateUserInput(name, email, password string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("Name is required")
	}
	if strings.TrimSpace(email) == "" {
		return errors.New("Email is required")
	}
	if !ValidateEmail(email) {
		return errors.New("Invalid email format")
	}
	if strings.TrimSpace(password) == "" {
		return errors.New("Password is required")
	}
	if len(password) < 6 {
		return errors.New("Password must be at least 6 characters")
	}
	return nil
}