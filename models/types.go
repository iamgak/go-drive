package models

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type User struct {
	ID              uint       `gorm:"primaryKey" json:"id" binding:"-"`
	Email           string     `gorm:"unique;not null" json:"email"`
	HashPassw       string     `gorm:"not null"`
	ActivationToken string     `gorm:"index"`
	Active          bool       `gorm:"default:false" json:"-"`
	VerifiedAt      time.Time  `gorm:"default:null"`
	CreatedAt       *time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP()" json:"created_at,omitempty" binding:"-"`
	UpdatedAt       *time.Time `gorm:"default:null" json:"-" binding:"-"`
}

type UserStruct struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	RepeatPassword string `json:"repeatPassword"`
}

type UsersSession struct {
	ID         uint       `gorm:"primaryKey"`
	UserID     uint       `gorm:"index"`
	LoginToken string     `gorm:"not null"`
	CreatedAt  *time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP()"`
}

type UserActivityLog struct {
	ID         uint       `gorm:"primaryKey"`
	UserID     uint       `gorm:"index"`
	Activity   string     `gorm:"not null"`
	Superseded bool       `gorm:"default:0"`
	IpAddr     string     `gorm:"default:null"`
	CreatedAt  *time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP()"`
	UpdatedAt  *time.Time `gorm:"default:null"`
}

type MyCustomClaims struct {
	Email  string `json:"email"`
	UserID uint   `json:"user_id"`
	jwt.StandardClaims
}
