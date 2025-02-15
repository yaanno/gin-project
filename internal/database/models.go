package database

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusLocked   UserStatus = "locked"
	UserStatusInactive UserStatus = "inactive"
	UserStatusDeleted  UserStatus = "deleted"
)

type User struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	Username       string     `json:"username" gorm:"unique;not null"`
	Email          string     `json:"email" gorm:"unique;not null"`
	Password       string     `json:"-" gorm:"not null"`
	Status         UserStatus `json:"status" gorm:"not null"`
	LastActivityAt time.Time  `json:"last_activity_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      time.Time  `json:"deleted_at,omitempty"`
	LockedUntil    time.Time  `json:"locked_until,omitempty"`
	LockReason     string     `json:"lock_reason,omitempty"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (u *User) HashPassword() error {
	// Use bcrypt with high cost for password hashing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost+4)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPasswordHash(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
