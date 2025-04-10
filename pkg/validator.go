package pkg

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Validator struct {
	Errors map[string]string
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddFieldError(key, message string) {
	if v.Errors == nil {
		v.Errors = make(map[string]string)
	}
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

func (v *Validator) NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func (v *Validator) MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

func (v *Validator) ValidEmail(email string) bool {
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailPattern.MatchString(email)
}

func (v *Validator) ValidPassword(password string) {
	v.CheckField(len(password) > 0, "password", "Password Should not be empty")
	if len(password) > 0 {
		emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%\-$@]{5,20}$`)
		v.CheckField(emailPattern.MatchString(password), "password", "Password Should contain Alphanumeric char and Special char (._%-$@) only between 5 to 20 char")
		// v.CheckField(len(password) < 20, "password", "Password Should be less than 20")
	}

}

func (f *Validator) ValidStatus(status string) bool {
	status = strings.ToLower(strings.Replace(status, "_", " ", 1))
	statusPattern := regexp.MustCompile(`^(pending|completed|in progress)$`)
	return statusPattern.MatchString(strings.ToLower(status))
}

func (app *Validator) ReadString(s string, defaultValue string) string {
	if s == "" {
		return defaultValue
	}
	return s
}

func (app *Validator) ReadInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}

	if i >= 50 {
		i = defaultValue
	}

	return i
}

func (f *Validator) ValidDate(input string) bool {
	input = strings.TrimSpace(input) // Remove leading/trailing spaces
	if input == "" {
		return false
	}

	_, err := time.Parse("2006-01-02", input)
	return err == nil
}

func (f *Validator) GetValidDate(date string) string {
	if f.ValidDate(date) {
		return date
	}
	return ""
}
