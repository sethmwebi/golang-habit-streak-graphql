package resolvers

import (
	"context"
	"fmt"
	"os"
	"time"

	"habit-streak-graphql/internal/middleware"
	"habit-streak-graphql/internal/models"
	"habit-streak-graphql/internal/utils"
)

// Register is the resolver for the register field.
func (r *mutationResolver) Register(ctx context.Context, name string, email string, password string) (*models.AuthPayload, error) {
	err := utils.ValidateName(name)
	if err != nil {
		return nil, fmt.Errorf("invalid name: %w", err)
	}

	err = utils.ValidateEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	err = utils.ValidatePasswordStrength(password)
	if err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := r.UserRepo.CreateUser(name, email, hashedPassword)
	if err != nil {
		return nil, err
	}

	token, err := utils.GenerateJWT(user.ID, os.Getenv("JWT_SECRET"))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.AuthPayload{
		Token: token,
		User:  user,
	}, nil
}

func (r *mutationResolver) Login(ctx context.Context, email string, password string) (*models.AuthPayload, error) {
	user, err := r.UserRepo.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	err = utils.ComparePassword(user.Password, password)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	token, err := utils.GenerateJWT(user.ID, os.Getenv("JWT_SECRET"))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.AuthPayload{
		Token: token,
		User:  user,
	}, nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, name *string, email *string, password *string) (*models.User, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("unauthorized")
	}

	if name != nil {
		if err := utils.ValidateName(*name); err != nil {
			return nil, fmt.Errorf("invalid name: %w", err)
		}
	}

	if email != nil {
		if err := utils.ValidateEmail(*email); err != nil {
			return nil, fmt.Errorf("invalid email: %w", err)
		}
	}

	var hashedPassword *string

	if password != nil {
		if err := utils.ValidatePasswordStrength(*password); err != nil {
			return nil, fmt.Errorf("invalid password: %w", err)
		}

		hash, err := utils.HashPassword(*password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}

		hashedPassword = &hash
	}

	user, err := r.UserRepo.UpdateUser(userID, name, email, hashedPassword)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *mutationResolver) DeleteUser(ctx context.Context) (bool, error) {
	userID, ok := middleware.GetUserID(ctx)

	if !ok {
		return false, fmt.Errorf("unauthorized")
	}

	deleted, err := r.UserRepo.DeleteUser(userID)
	if err != nil {
		return false, err
	}

	if !deleted {
		return false, fmt.Errorf("user not found")
	}
	return true, nil
}

func (r *mutationResolver) CreateHabit(ctx context.Context, name string, description string) (*models.Habit, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("unauthorized")
	}

	if err := utils.ValidateName(name); err != nil {
		return nil, fmt.Errorf("invalid habit name: %w", err)
	}

	if err := utils.ValidateDescription(description); err != nil {
		return nil, fmt.Errorf("invalid description: %w", err)
	}

	habit, err := r.HabitRepo.CreateHabit(userID, name, description)
	if err != nil {
		return nil, err
	}

	return habit, err
}

func (r *mutationResolver) UpdateHabit(ctx context.Context, id string, name *string, description *string) (*models.Habit, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("unauthorized")
	}

	if name != nil {
		if err := utils.ValidateName(*name); err != nil {
			return nil, fmt.Errorf("invalid name: %w", err)
		}
	}

	if description != nil {
		if err := utils.ValidateDescription(*description); err != nil {
			return nil, fmt.Errorf("invalid description: %w", err)
		}
	}

	habit, err := r.HabitRepo.UpdateHabit(id, userID, name, description)
	if err != nil {
		return nil, err
	}
	return habit, nil
}

func (r *mutationResolver) DeleteHabit(ctx context.Context, id string) (bool, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return false, fmt.Errorf("unauthorized")
	}
	deleted, err := r.HabitRepo.DeleteHabit(id, userID)
	if err != nil {
		return false, err
	}
	if !deleted {
		return false, fmt.Errorf("habit not found")
	}
	return true, nil
}

func (r *mutationResolver) CheckInHabit(ctx context.Context, habitID string, date *string) (*models.HabitLog, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("unauthorized")
	}

	var completedDate time.Time

	if date == nil {
		completedDate = time.Now().UTC()
	} else {
		var err error
		completedDate, err = time.Parse("2006-01-02", *date)
		if err != nil {
			return nil, fmt.Errorf("invalid date format - YYYY-MM-DD")
		}
	}

	_, err := r.HabitRepo.GetHabitWithUserCheck(habitID, userID)
	if err != nil {
		return nil, err
	}

	isDuplicate, err := r.HabitLogRepo.CheckDuplicateLog(habitID, completedDate)
	if err != nil {
		return nil, err
	}

	if isDuplicate {
		return nil, fmt.Errorf("you have already checked in for this habit today")
	}

	log, err := r.HabitLogRepo.CreateHabitLog(habitID, completedDate)
	if err != nil {
		return nil, err
	}
	return log, nil
}

func (r *mutationResolver) DeleteHabitLog(ctx context.Context, id string) (bool, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return false, fmt.Errorf("unauthorized")
	}

	log, err := r.HabitLogRepo.GetHabitLogByID(id)
	if err != nil {
		return false, nil
	}

	if log == nil {
		return false, fmt.Errorf("habit log not found")
	}

	habit, err := r.HabitRepo.GetHabitWithUserCheck(log.HabitId, userID)

	if habit == nil {
		return false, fmt.Errorf("parent habit not found")
	}

	if habit.UserID != userID {
		return false, fmt.Errorf("unauthorized")
	}
	deleted, err := r.HabitLogRepo.DeleteHabitLog(id)
	if err != nil {
		return false, err
	}

	if !deleted {
		return false, fmt.Errorf("habit log not found")
	}
	return true, nil
}
