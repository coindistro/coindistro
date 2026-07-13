package validation

import (
	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground validator.
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator instance.
func New() *Validator {
	v := validator.New()

	// Register custom validations
	_ = v.RegisterValidation("password", validatePassword)
	_ = v.RegisterValidation("crypto_address", validateCryptoAddress)
	_ = v.RegisterValidation("phone", validatePhone)

	return &Validator{validate: v}
}

// Validate validates a struct and returns validation errors.
func (v *Validator) Validate(i interface{}) error {
	return v.validate.Struct(i)
}

// ValidateVar validates a single variable.
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

// RegisterValidation registers a custom validation function.
func (v *Validator) RegisterValidation(tag string, fn validator.Func) error {
	return v.validate.RegisterValidation(tag, fn)
}

// validatePassword validates password strength.
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		default:
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// validateCryptoAddress validates a basic crypto address format.
func validateCryptoAddress(fl validator.FieldLevel) bool {
	address := fl.Field().String()
	if len(address) < 26 || len(address) > 64 {
		return false
	}

	for _, char := range address {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}

	return true
}

// validatePhone validates a phone number format.
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if len(phone) < 10 || len(phone) > 15 {
		return false
	}

	for _, char := range phone {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}
