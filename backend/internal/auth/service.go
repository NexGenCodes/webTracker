package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/database/db"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/notif"
)

// Service handles authentication business logic
type Service struct {
	cfg     *config.Config
	queries *db.Queries
	mailer  *notif.Mailer
}

// NewService creates a new auth service
func NewService(cfg *config.Config, queries *db.Queries) *Service {
	return &Service{
		cfg:     cfg,
		queries: queries,
		mailer:  notif.NewMailer(cfg),
	}
}

// GenerateOTP creates a 6 digit code and sends it via email, returning the stateless token
func (s *Service) GenerateOTP(ctx context.Context, companyName, email, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	// Check if already registered
	company, err := s.queries.GetCompanyByEmail(ctx, email)
	if err == nil && company.ID != uuid.Nil {
		return "", errors.New("a company with this email already exists")
	}

	// Generate 6 digit OTP
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	otp := fmt.Sprintf("%06d", r.Intn(1000000))
	logger.Info().Str("email", email).Msg("Generated OTP, sending verification email")
	s.mailer.SendAsync(notif.OTPEmail(email, otp))

	// Hash password and OTP
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	hashedOTP, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash OTP: %w", err)
	}

	// Create Stateless JWT Token
	claims := OTPClaims{
		CompanyName:    companyName,
		Email:          email,
		HashedPassword: string(hashedPassword),
		HashedOTP:      string(hashedOTP),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "webtracker-auth-otp",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// VerifyOTP validates the OTP from the token and creates the user record
func (s *Service) VerifyOTP(ctx context.Context, otp string, otpToken string) (*AuthResponse, string, error) {
	// Parse OTP token
	claims := &OTPClaims{}
	token, err := jwt.ParseWithClaims(otpToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, "", errors.New("invalid or expired OTP session")
	}

	// Verify OTP
	err = bcrypt.CompareHashAndPassword([]byte(claims.HashedOTP), []byte(otp))
	if err != nil {
		return nil, "", errors.New("incorrect OTP code")
	}

	// Create company in DB
	company, err := s.queries.CreateCompany(ctx, db.CreateCompanyParams{
		AdminEmail: claims.Email,
		Name:       sql.NullString{String: claims.CompanyName, Valid: true},
		SetupToken: sql.NullString{String: uuid.New().String(), Valid: true},
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create company: %w", err)
	}

	// Update password hash immediately
	err = s.queries.SetCompanyPassword(ctx, db.SetCompanyPasswordParams{
		ID:                company.ID,
		AdminPasswordHash: sql.NullString{String: claims.HashedPassword, Valid: true},
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to set company password: %w", err)
	}

	// Generate Session JWT
	sessionToken, err := s.generateJWT(company.ID, company.Name.String, company.AdminEmail, company.PlanType.String, "pending_verification")
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	return &AuthResponse{
		CompanyID:  company.ID,
		Email:      company.AdminEmail,
		PlanType:   company.PlanType.String,
		AuthStatus: "pending_verification",
	}, sessionToken, nil
}

// SetupCompany finalizes onboarding (WhatsApp, Name, Prefix)
func (s *Service) SetupCompany(ctx context.Context, companyID uuid.UUID, req SetupCompanyRequest) (*AuthResponse, string, error) {
	prefix := strings.ToUpper(strings.TrimSpace(req.TrackingPrefix))

	// Re-fetch company to get the name for abbreviation if prefix is empty
	company, _ := s.queries.GetCompanyByID(ctx, companyID)

	if prefix == "" {
		prefix = config.GenerateAbbreviation(company.Name.String)
	}

	err := s.queries.UpdateCompanyOnboarding(ctx, db.UpdateCompanyOnboardingParams{
		ID:             companyID,
		WhatsappPhone:  sql.NullString{String: req.WhatsappPhone, Valid: true},
		TrackingPrefix: sql.NullString{String: prefix, Valid: true},
	})
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return nil, "", errors.New("this tracking prefix is already in use")
		}
		return nil, "", fmt.Errorf("failed to save setup data: %w", err)
	}

	// Generate updated JWT with active status
	sessionToken, err := s.generateJWT(company.ID, company.Name.String, company.AdminEmail, company.PlanType.String, company.AuthStatus.String)

	return &AuthResponse{
		CompanyID:   company.ID,
		CompanyName: company.Name.String,
		Email:       company.AdminEmail,
		PlanType:    company.PlanType.String,
		AuthStatus:  company.AuthStatus.String,
	}, sessionToken, nil
}

// Login verifies credentials and returns a JWT
func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, string, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	company, err := s.queries.GetCompanyByEmail(ctx, email)
	if err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	if !company.AdminPasswordHash.Valid {
		return nil, "", errors.New("account not fully set up")
	}

	err = bcrypt.CompareHashAndPassword([]byte(company.AdminPasswordHash.String), []byte(req.Password))
	if err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	token, err := s.generateJWT(company.ID, company.Name.String, company.AdminEmail, company.PlanType.String, company.AuthStatus.String)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	return &AuthResponse{
		CompanyID:   company.ID,
		CompanyName: company.Name.String,
		Email:       company.AdminEmail,
		PlanType:    company.PlanType.String,
		AuthStatus:  company.AuthStatus.String,
	}, token, nil
}

func (s *Service) generateJWT(companyID uuid.UUID, companyName, email, planType, authStatus string) (string, error) {
	if s.cfg.JWTSecret == "" {
		return "", errors.New("JWT_SECRET is not configured")
	}

	claims := JWTClaims{
		CompanyID:   companyID,
		CompanyName: companyName,
		Email:       email,
		PlanType:    planType,
		AuthStatus:  authStatus,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "webtracker-auth",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
