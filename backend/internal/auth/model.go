package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// LoginRequest payload for /api/auth/login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AdminLoginRequest payload for /api/auth/admin-login
type AdminLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RegisterIntentRequest payload for /api/auth/register-intent
type RegisterIntentRequest struct {
	CompanyName string `json:"company_name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
}

// VerifyOTPRequest payload for /api/auth/verify-otp
type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6"`
}

// SetupCompanyRequest payload for /api/auth/setup
type SetupCompanyRequest struct {
	WhatsappPhone  string `json:"whatsapp_phone" validate:"required"`
	TrackingPrefix string `json:"tracking_prefix" validate:"omitempty,max=5"`
}

type AuthResponse struct {
	CompanyID   uuid.UUID `json:"company_id"`
	CompanyName string    `json:"company_name"`
	Email       string    `json:"email"`
	PlanType    string    `json:"plan_type"`
	AuthStatus  string    `json:"auth_status"`
	Token       string    `json:"token,omitempty"`
}

// JWTClaims custom claims for our main JWT tokens
type JWTClaims struct {
	CompanyID   uuid.UUID `json:"company_id"`
	Email       string    `json:"email"`
	CompanyName string    `json:"company_name"`
	PlanType    string    `json:"plan_type"`
	AuthStatus  string    `json:"auth_status"`
	AppRole     string    `json:"app_role"`
	Role        string    `json:"role"`
	jwt.RegisteredClaims
}

// PendingUserPayload is the JSON stored in Redis for registration
type PendingUserPayload struct {
	CompanyName    string `json:"company_name"`
	Email          string `json:"email"`
	HashedOTP      string `json:"hashed_otp"`
	HashedPassword string `json:"hashed_password"`
}

// ForgotPasswordRequest payload for /api/auth/forgot-password
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest payload for /api/auth/reset-password
type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	OTP         string `json:"otp" validate:"required,len=6"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// PendingResetPayload is the JSON stored in Redis for password reset
type PendingResetPayload struct {
	Email     string `json:"email"`
	HashedOTP string `json:"hashed_otp"`
}

