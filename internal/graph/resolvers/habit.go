package resolvers

import (
	"context"
	"fmt"
	"time"

	"habit-streak-graphql/internal/models"
	"habit-streak-graphql/internal/utils"
)

func (r *habitResolver) CreatedAt(ctx context.Context, obj *models.Habit) (string, error) {
	return obj.CreatedAt.Format(time.RFC3339), nil
}

// UpdatedAt is the resolver for the updatedAt field.
func (r *habitResolver) UpdatedAt(ctx context.Context, obj *models.Habit) (string, error) {
	return obj.UpdatedAt.Format(time.RFC3339), nil
}

func (r *habitResolver) User(ctx context.Context, obj *models.Habit) (*models.User, error) {
	user, err := r.UserRepo.GetUserById(obj.UserID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (r *habitResolver) Logs(ctx context.Context, obj *models.Habit) ([]*models.HabitLog, error) {
	logs, err := r.HabitLogRepo.GetHabitLogsByHabitID(obj.ID)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *habitResolver) TotalCompletions(ctx context.Context, obj *models.Habit) (int32, error) {
	count, err := r.HabitLogRepo.CountTotalCompletions(obj.ID)
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}

func (r *habitResolver) CurrentStreak(ctx context.Context, obj *models.Habit) (int32, error) {
	logs, err := r.HabitLogRepo.GetHabitLogsByHabitID(obj.ID)
	if err != nil {
		return 0, nil
	}
	count := utils.CalculateCurrentStreak(logs)
	return int32(count), nil
}

// LongestStreak is the resolver for the longestStreak field.
func (r *habitResolver) LongestStreak(ctx context.Context, obj *models.Habit) (int32, error) {
	logs, err := r.HabitLogRepo.GetHabitLogsByHabitID(obj.ID)
	if err != nil {
		return 0, nil
	}
	count := utils.CalculateLongestStreak(logs)
	return int32(count), nil
}
