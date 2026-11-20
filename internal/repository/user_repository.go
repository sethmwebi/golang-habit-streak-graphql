package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"habit-streak-graphql/internal/models"

	"github.com/lib/pq"
)

// connection pool -> a collection of db connections
type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

func (r *UserRepository) CreateUser(name string, email string, hashedPassword string) (*models.User, error) {
	now := time.Now()

	var user models.User
	err := r.DB.QueryRow(`INSERT INTO users(id, name, email, password, created_at, updated_at) VALUES(gen_random_uuid(), $1, $2, $3, $4, $5) RETURNING id, name, email, password, created_at, updated_at`, name, email, hashedPassword, now, now).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("email already registered")
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.DB.QueryRow("SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = $1", email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invalid email or password")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetUserById(id string) (*models.User, error) {
	var user models.User
	err := r.DB.QueryRow("SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1", id).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}

		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) UpdateUser(id string, name *string, email *string, password *string) (*models.User, error) {
	var setClauses []string
	var args []any

	if name != nil {
		args = append(args, *name)
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", len(args)))
	}

	if email != nil {
		args = append(args, *email)
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", len(args)))
	}

	if password != nil {
		args = append(args, *password)
		setClauses = append(setClauses, fmt.Sprintf("password = $%d", len(args)))
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("no fields provided to update")
	}

	args = append(args, time.Now())
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", len(args)))

	args = append(args, id)
	idPlaceholder := len(args)

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d RETURNING id, name, email, created_at, updated_at", strings.Join(setClauses, ", "), idPlaceholder)

	var user models.User
	err := r.DB.QueryRow(query, args...).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}

		if isUniqueViolation(err) {
			return nil, fmt.Errorf("email already in use")
		}

		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) DeleteUser(id string) (bool, error) {
	result, err := r.DB.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return false, fmt.Errorf("failed to delete user: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to confirm deletion: %w", err)
	}

	return rowsAffected > 0, nil
}
