package database

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusLocked   UserStatus = "locked"
	UserStatusInactive UserStatus = "inactive"
	UserStatusDeleted  UserStatus = "deleted"
)

type User struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	Username       string         `gorm:"unique;not null;size:100" json:"username"`
	Email          string         `gorm:"unique;not null;size:100" json:"email"`
	Password       string         `gorm:"not null" json:"-"`
	Status         UserStatus     `gorm:"not null;default:'active'" json:"status"`
	LastActivityAt time.Time      `gorm:"default:null" json:"last_activity_at"`
	LockedUntil    time.Time      `gorm:"default:null" json:"locked_until,omitempty"`
	LockReason     string         `gorm:"default:null" json:"lock_reason,omitempty"`
	CreatedAt      time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"default:null" json:"deleted_at,omitempty"`
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

type LoginAttempt struct {
	gorm.Model
	Username    string    `gorm:"not null;index:idx_username_ip,unique" json:"username"`
	IpAddress   string    `gorm:"not null;index:idx_username_ip,unique" json:"ip_address"`
	Attempts    uint      `gorm:"not null;default:0" json:"attempts"`
	Success     bool      `gorm:"not null" json:"success"`
	LastAttempt time.Time `gorm:"default:null" json:"last_attempt"`
}
