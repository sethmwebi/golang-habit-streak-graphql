package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"habit-streak-graphql/internal/models"
)

type HabitLogRepository struct {
	DB *sql.DB
}

func NewHabitLogRepository(db *sql.DB) *HabitLogRepository {
	return &HabitLogRepository{DB: db}
}

func toDateString(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

func (r *HabitLogRepository) CheckDuplicateLog(habitID string, completedDate time.Time) (bool, error) {
	dateStr := toDateString(completedDate)

	var exists bool
	err := r.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM habit_logs WHERE habit_id = $1 AND completed_date = $2::date)", habitID, dateStr).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check duplicate log: %w", err)
	}
	return exists, nil
}

func (r *HabitLogRepository) CreateHabitLog(habitID string, completedDate time.Time) (*models.HabitLog, error) {
	now := time.Now()
	dateStr := toDateString(completedDate)

	var log models.HabitLog
	err := r.DB.QueryRow(`INSERT INTO habit_logs(id, habit_id, completed_date, created_at) VALUES (gen_random_uuid()::text, $1, $2::date, $3) RETURNING id, habit_id, completed_date, created_at`, habitID, dateStr, now).Scan(&log.ID, &log.HabitId, &log.CompletedDate, &log.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("you have already checked in for this habit today or on this date")
		}
		return nil, fmt.Errorf("failed to create habit log: %w", err)
	}
	return &log, nil
}

func (r *HabitLogRepository) GetHabitLogsByHabitID(habitID string) ([]*models.HabitLog, error) {
	rows, err := r.DB.Query("SELECT id, habit_id, completed_date, created_at FROM habit_logs WHERE habit_id = $1 ORDER BY completed_date DESC", habitID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch habit logs: %w", err)
	}

	defer rows.Close()

	logs := make([]*models.HabitLog, 0)

	for rows.Next() {
		var log models.HabitLog
		err := rows.Scan(&log.ID, &log.HabitId, &log.CompletedDate, &log.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan habit log: %w", err)
		}

		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating habit logs: %w", err)
	}
	return logs, nil
}

func (r *HabitLogRepository) GetHabitLogByID(id string) (*models.HabitLog, error) {
	var log models.HabitLog
	err := r.DB.QueryRow("SELECT id, habit_id, completed_date, created_at FROM habit_logs WHERE id = $1", id).Scan(&log.ID, &log.HabitId, &log.CompletedDate, &log.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch habit log: %w", err)
	}
	return &log, nil
}

func (r *HabitLogRepository) DeleteHabitLog(id string) (bool, error) {
	result, err := r.DB.Exec("DELETE FROM habit_logs WHERE id = $1", id)
	if err != nil {
		return false, fmt.Errorf("failed to delete habit log: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to confirm deletion: %w", err)
	}

	return rowsAffected > 0, nil
}

func (r *HabitLogRepository) CountTotalCompletions(habitID string) (int, error) {
	var count int
	err := r.DB.QueryRow("SELECT COUNT(*) FROM habit_logs WHERE habit_id = $1", habitID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count completions: %w", err)
	}
	return count, nil
}
