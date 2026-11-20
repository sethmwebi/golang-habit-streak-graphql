package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"habit-streak-graphql/internal/models"
)

type HabitRepository struct {
	DB *sql.DB
}

func NewHabitRepository(db *sql.DB) *HabitRepository {
	return &HabitRepository{DB: db}
}

func (r *HabitRepository) CreateHabit(userID string, name string, description string) (*models.Habit, error) {
	now := time.Now()

	// Postgres generates the ID and returns the new row in one statement
	var habit models.Habit
	err := r.DB.QueryRow(`INSERT INTO habits(id, user_id, name, description, created_at, updated_at) VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5) RETURNING id, user_id, name, description, created_at, updated_at`, userID, name, description, now, now).Scan(&habit.ID, &habit.UserID, &habit.Name, &habit.Description, &habit.CreatedAt, &habit.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create this habit: %w", err)
	}

	return &habit, nil
}

func (r *HabitRepository) GetHabitsByUserID(userID string) ([]*models.Habit, error) {
	// *sql.Rows => cursor
	rows, err := r.DB.Query("SELECT id, user_id, name, description, created_at, updated_at FROM habits WHERE user_id = $1 ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch habits: %w", err)
	}

	defer rows.Close()

	habits := make([]*models.Habit, 0)

	for rows.Next() {
		var habit models.Habit
		err := rows.Scan(
			&habit.ID,
			&habit.UserID,
			&habit.Name,
			&habit.Description,
			&habit.CreatedAt,
			&habit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan habit: %w", err)
		}

		habits = append(habits, &habit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating habits: %w", err)
	}

	return habits, nil
}

func (r *HabitRepository) GetHabitWithUserCheck(habitID string, userID string) (*models.Habit, error) {
	var habit models.Habit

	err := r.DB.QueryRow("SELECT id, user_id, name, description, created_at, updated_at FROM habits WHERE id = $1 AND user_id = $2", habitID, userID).Scan(&habit.ID, &habit.UserID, &habit.Name, &habit.Description, &habit.CreatedAt, &habit.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("habit not found")
		}

		return nil, fmt.Errorf("failed to fetch habit: %w", err)
	}
	return &habit, nil
}

func (r *HabitRepository) UpdateHabit(habitID string, userID string, name *string, description *string) (*models.Habit, error) {
	var setClauses []string
	var args []any

	// Postgres uses numbered placeholders, so each number comes from len(args)
	if name != nil {
		args = append(args, *name)
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", len(args)))
	}

	if description != nil {
		args = append(args, *description)
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", len(args)))
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("no fields provided to update")
	}

	args = append(args, time.Now())
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", len(args)))

	args = append(args, habitID)
	habitIDPlaceholder := len(args)
	args = append(args, userID)
	userIDPlaceholder := len(args)

	query := fmt.Sprintf(`UPDATE habits SET %s WHERE id = $%d AND user_id = $%d RETURNING id, user_id, name, description, created_at, updated_at`, strings.Join(setClauses, ", "), habitIDPlaceholder, userIDPlaceholder)

	var habit models.Habit
	err := r.DB.QueryRow(query, args...).Scan(&habit.ID, &habit.UserID, &habit.Name, &habit.Description, &habit.CreatedAt, &habit.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("habit not found")
		}
		return nil, fmt.Errorf("failed to update habit: %w", err)
	}

	return &habit, nil
}

func (r *HabitRepository) DeleteHabit(habitID string, userID string) (bool, error) {
	result, err := r.DB.Exec("DELETE FROM habits WHERE id = $1 AND user_id = $2", habitID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to delete habit: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to delete habit: %w", err)
	}

	if rowsAffected == 0 {
		return false, fmt.Errorf("habit not found")
	}

	return true, nil
}

func (r *HabitRepository) GetHabitByID(id string) (*models.Habit, error) {
	var habit models.Habit

	err := r.DB.QueryRow("SELECT id, user_id, name, description, created_at, updated_at FROM habits WHERE id = $1", id).Scan(&habit.ID, &habit.UserID, &habit.Name, &habit.Description, &habit.CreatedAt, &habit.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch habit: %w", err)
	}
	return &habit, nil
}
