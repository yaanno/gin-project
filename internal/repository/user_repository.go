package repository

import (
	"time"

	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/internal/database"
	"github.com/yourusername/user-management-api/pkg/errors/apperrors"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *database.User) error
	FindUserByUsername(username string) (*database.User, error)
	FindUserByID(userID uint) (*database.User, error)
	UpdateUser(user *database.User) error
	DeleteUser(userID uint) error
	GetAllUsers() ([]database.User, error)
	LockUser(userID uint, reason string, duration time.Duration) error
	UnlockUser(userID uint) error
	MarkUserInactive(userID uint) error
	HardDeleteUser(userID uint) error
	HardDeletePermanentlyInactiveUsers() error
	LockSecurityViolationUsers() error
	MarkInactiveUsers() error
}

type UserRepositoryImpl struct {
	db  *gorm.DB
	log zerolog.Logger
}

func NewUserRepository(db *gorm.DB, log zerolog.Logger) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		db:  db,
		log: log.With().Str("repository", "UserRepository").Logger(),
	}
}

func (r *UserRepositoryImpl) CreateUser(user *database.User) error {
	result := r.db.Create(user)
	if result.Error != nil {
		r.log.Error().Err(result.Error).Str("user", user.Username).Msg("Failed to create user")
		return apperrors.NewDatabaseError("Failed to create user", result.Error)
	}
	return nil
}

func (r *UserRepositoryImpl) FindUserByUsername(username string) (*database.User, error) {
	user := &database.User{}
	result := r.db.First(user, "username = ?", username)
	if result.Error == gorm.ErrRecordNotFound {
		r.log.Error().Err(result.Error).Str("username", username).Msg("User not found")
		return &database.User{}, apperrors.NewNotFoundError("User not found", result.Error, "username", username)
	}

	if result.Error != nil {
		r.log.Error().Err(result.Error).Str("username", username).Msg("Failed to find user")
		return &database.User{}, apperrors.NewDatabaseError("Failed to find user", result.Error)
	}

	return user, nil
}

func (r *UserRepositoryImpl) FindUserByID(userID uint) (*database.User, error) {
	user := &database.User{}
	result := r.db.First(user, "id = ?", userID)

	if result.Error == gorm.ErrRecordNotFound {
		r.log.Error().Err(result.Error).Uint("user_id", userID).Msg("User not found")
		return &database.User{}, apperrors.NewNotFoundError("User not found", result.Error, "id", userID)
	}

	if result.Error != nil {
		r.log.Error().Err(result.Error).Uint("user_id", userID).Msg("Failed to find user")
		return &database.User{}, apperrors.NewDatabaseError("Failed to find user", result.Error)
	}

	return user, nil
}

func (r *UserRepositoryImpl) UpdateUser(user *database.User) error {
	result := r.db.Updates(user)

	if result.Error != nil {
		r.log.Error().Err(result.Error).Str("user", user.Username).Msg("Failed to update user")
		return apperrors.NewDatabaseError("Failed to update user", result.Error)
	}
	return nil
}

func (r *UserRepositoryImpl) DeleteUser(userID uint) error {
	result := r.db.Where("id = ?", userID).Updates(&database.User{
		Status:    database.UserStatusDeleted,
		DeletedAt: gorm.DeletedAt{Valid: true},
	})
	if result.Error != nil {
		r.log.Error().Err(result.Error).Uint("user_id", userID).Msg("Failed to delete user")
		return apperrors.NewDatabaseError("Failed to delete user", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewNotFoundError("User not found", nil, "id", userID)
	}
	return nil
}

func (r *UserRepositoryImpl) GetAllUsers() ([]database.User, error) {
	var users []database.User
	result := r.db.Find(&users)
	if result.Error != nil {
		r.log.Error().Err(result.Error).Msg("Failed to get users")
		return []database.User{}, apperrors.NewDatabaseError("Failed to get users", result.Error)
	}
	return users, nil
}

func (r *UserRepositoryImpl) LockUser(userID uint, reason string, duration time.Duration) error {
	lockedUntil := time.Now().Add(duration)
	result := r.db.Updates(&database.User{
		ID:          userID,
		Status:      database.UserStatusLocked,
		LockReason:  reason,
		LockedUntil: lockedUntil,
	})
	if result.Error != nil {
		r.log.Error().Err(result.Error).Uint("user_id", userID).Msg("Failed to lock user")
		return apperrors.NewDatabaseError("Failed to lock user", result.Error)
	}
	return nil
}

func (r *UserRepositoryImpl) UnlockUser(userID uint) error {
	result := r.db.Updates(&database.User{
		ID:         userID,
		Status:     database.UserStatusActive,
		LockReason: "",
	})
	if result.Error != nil {
		r.log.Error().Err(result.Error).Uint("user_id", userID).Msg("Failed to unlock user")
		return apperrors.NewDatabaseError("Failed to unlock user", result.Error)
	}
	return nil
}

func (r *UserRepositoryImpl) MarkUserInactive(userID uint) error {
	r.log.Info().Uint("user_id", userID).Msg("Marking user inactive")
	result := r.db.Updates(&database.User{
		ID:     userID,
		Status: database.UserStatusInactive,
	})
	if result.Error != nil {
		r.log.Error().Err(result.Error).Uint("user_id", userID).Msg("Failed to mark user inactive")
		return apperrors.NewDatabaseError("Failed to mark user inactive", result.Error)
	}
	r.log.Info().Uint("user_id", userID).Msg("User marked inactive")
	return nil
}

func (r *UserRepositoryImpl) HardDeleteUser(userID uint) error {
	result := r.db.Delete(&database.User{}, "id = ?", userID)
	if result.Error != nil {
		r.log.Error().Err(result.Error).Uint("user_id", userID).Msg("Failed to hard delete user")
		return apperrors.NewDatabaseError("Failed to hard delete user", result.Error)
	}
	return nil
}

func (r *UserRepositoryImpl) HardDeleteUserMarkedForDeletion(userID uint) error {
	result := r.db.Unscoped().Where("id = ? AND status = ?", userID, database.UserStatusDeleted).Delete(&database.User{})
	if result.Error != nil {
		r.log.Error().Err(result.Error).Uint("user_id", userID).Msg("Failed to hard delete user")
		return apperrors.NewDatabaseError("Failed to hard delete user", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewNotFoundError("User not found", nil, "id", userID)
	}
	return nil
}

func (r *UserRepositoryImpl) HardDeletePermanentlyInactiveUsers() error {
	result := r.db.Unscoped().Where(
		"status = ? AND deleted_at < ?",
		database.UserStatusInactive,
		time.Now().AddDate(0, 0, -365),
	).Delete(&database.User{})

	if result.Error != nil {
		r.log.Error().Err(result.Error).Msg("Failed to hard delete permanently inactive users")
		return apperrors.NewDatabaseError("Failed to hard delete permanently inactive users", result.Error)
	}

	r.log.Info().
		Int64("deleted_users", result.RowsAffected).
		Msg("Deleted permanently inactive users")

	return nil
}

func (r *UserRepositoryImpl) MarkInactiveUsers() error {
	result := r.db.Model(&database.User{}).
		Where("status = ? AND last_activity_at < ?",
			database.UserStatusActive,
			time.Now().AddDate(0, 0, -90),
		).
		Updates(map[string]interface{}{
			"status":     database.UserStatusInactive,
			"deleted_at": time.Now(),
		})

	if result.Error != nil {
		r.log.Error().Err(result.Error).Msg("Failed to mark inactive")
		return apperrors.NewDatabaseError("Failed to mark inactive", result.Error)
	}

	r.log.Info().
		Int64("marked_inactive", result.RowsAffected).
		Msg("Marked inactive users")

	return nil
}
func (r *UserRepositoryImpl) LockSecurityViolationUsers() error {
	subQuery := r.db.Table("login_attempts").
		Select("DISTINCT username").
		Where("success = ? AND last_attempt > ?",
			false,
			time.Now().AddDate(0, 0, -30),
		)

	result := r.db.Model(&database.User{}).
		Where("username IN (?)", subQuery).
		Updates(map[string]interface{}{
			"status":       database.UserStatusLocked,
			"locked_until": time.Now().Add(24 * time.Hour),
			"lock_reason":  "Multiple failed login attempts",
		})

	if result.Error != nil {
		r.log.Error().Err(result.Error).Msg("Failed to lock security violation users")
		return apperrors.NewDatabaseError("Failed to lock security violation users", result.Error)
	}

	r.log.Info().
		Int64("locked_users", result.RowsAffected).
		Msg("Locked users with security violations")

	return nil
}

var _ UserRepository = (*UserRepositoryImpl)(nil)
