package models

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/iamgak/go-drive/pkg"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserModelORM struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func (m *UserModelORM) RegisterUser(ctx context.Context, email, password, ip string) error {
	hashedPassword, err := m.GeneratePassword(password)
	if err != nil {
		return err
	}
	token := m.GenerateSHA1Hash(ip)
	user := User{Email: email, HashPassw: string(hashedPassword), ActivationToken: token}
	result := m.db.WithContext(ctx).Create(&user)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return pkg.ErrNoRecord
	}

	activity := UserActivityLog{UserID: user.ID, Activity: "New User Register"}
	return m.UserActivityLog(&activity)
}

func (m *UserModelORM) LoginUser(c context.Context, creds *UserStruct) (string, error) {
	var user User
	if err := m.db.WithContext(c).Where("email = ?", strings.TrimSpace(creds.Email)).First(&user).Error; err != nil {
		m.logger.Error("Error fetching data", err)
		return "", pkg.ErrInvalidCredentials
	}

	if !user.Active {
		return "", pkg.ErrAccountInActive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashPassw), []byte(creds.Password)); err != nil {
		m.logger.Error("Error handling passw", err)
		return "", pkg.ErrInvalidCredentials
	}

	token, err := m.generateToken(user.Email, user.ID)
	if err != nil {
		return "", err
	}

	activity := UserActivityLog{UserID: user.ID, Activity: "Logged In"}
	err = m.UserActivityLog(&activity)
	if err != nil {
		return "", err
	}

	return token, m.CreateSession(token, user.ID)
}

func (m *UserModelORM) GeneratePassword(newPassword string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(newPassword), 12)
}

func (m *UserModelORM) CreateSession(token string, user_id uint) error {
	user_session := UsersSession{UserID: user_id, LoginToken: token}
	return m.db.Create(&user_session).Error
}
func (m *UserModelORM) ActivateAccount(token string) error {
	var user User
	if err := m.db.Select("id").Where("activation_token = ?", token).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.ErrNoRecord
		}
		return err
	}

	result := m.db.Model(&user).Updates(map[string]any{
		"activation_token": nil,
		"active":           true,
		"verified_at":      time.Now(),
	})

	if result.Error != nil {
		return result.Error
	}

	activity := UserActivityLog{UserID: user.ID, Activity: "Account Activated"}
	return m.UserActivityLog(&activity)
}

func (m *UserModelORM) generateToken(email string, userID uint) (string, error) {
	if err := godotenv.Load(); err != nil {
		m.logger.Error(err.Error())
		return "", pkg.ErrInternalServer
	}

	signingKey := []byte(os.Getenv("SIGNING_KEY"))
	claims := MyCustomClaims{
		Email:  email,
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(4 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey)
}

func (m *UserModelORM) emailExists(email string) bool {
	var count int64
	m.db.Model(&User{}).Where("email = ?", email).Count(&count)
	return count > 0
}

func (m *UserModelORM) GenerateSHA1Hash(ip string) string {
	privateKey := "hello_world124#@##$$2"
	rand.Seed(time.Now().UnixNano())
	randomInt := rand.Intn(1000000)
	input := fmt.Sprintf("%s:%d:%s", ip, randomInt, privateKey)
	hash := sha1.New()
	hash.Write([]byte(input))
	return hex.EncodeToString(hash.Sum(nil))
}

func (m *UserModelORM) ValidateUserData(user *UserStruct, register bool) *pkg.Validator {
	validator := &pkg.Validator{
		Errors: make(map[string]string),
	}

	validator.CheckField(validator.NotBlank(user.Email), "email", "Please, fill the email field")
	validator.CheckField(validator.NotBlank(user.Password), "password", "Please, fill the password field")
	validator.ValidPassword(user.Password)

	if validator.Errors["email"] == "" {
		validator.CheckField(validator.ValidEmail(user.Email), "email", "Invalid Email Format")
	}

	if register {
		validator.CheckField(validator.NotBlank(user.RepeatPassword), "repeatPassword", "Please, fill the repeat password field")
		if m.emailExists(user.Email) {
			validator.CheckField(false, "email", "Email already registered")
		}
		if user.Password != user.RepeatPassword {
			validator.Errors["repeatPassword"] = "Password not matched"
		}

	}

	return validator
}

func (m *UserModelORM) UserActivityLog(activity *UserActivityLog) error {
	result := m.db.Model(&UserActivityLog{}).Where("user_id = ? AND activity =? AND superseded = 0", activity.UserID, activity.Activity).
		Updates(map[string]any{"superseded": 1, "updated_at": time.Now()})

	if result.Error != nil && result.Error != sql.ErrNoRows {
		return result.Error
	}

	return m.db.Create(&activity).Error
}
