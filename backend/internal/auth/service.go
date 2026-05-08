package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"encoding/json"

	"webtracker-bot/internal/cache"
	"webtracker-bot/internal/config"
	"webtracker-bot/internal/database/db"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/notif"
	"webtracker-bot/internal/utils"
)

// defaultJWTKeyID is used when JWT_KEY_ID is not set in the environment.
const defaultJWTKeyID = "3ac00c7e-2058-4c54-8cf1-54ebca7a67f1"

// Service handles authentication business logic
type Service struct {
	cfg          *config.Config
	queries      db.Querier
	mailer       *notif.Mailer
	privateKey   *ecdsa.PrivateKey // Cached at init to avoid per-request disk reads
	loginLimiter *cache.LoginFailLimiter
	jwtKeyID     string // Key ID embedded in JWT headers — from JWT_KEY_ID env var
}

// NewService creates a new auth service
func NewService(cfg *config.Config, queries db.Querier) *Service {
	kid := cfg.JWTKeyID
	if kid == "" {
		kid = defaultJWTKeyID
	}
	s := &Service{
		cfg:          cfg,
		queries:      queries,
		mailer:       notif.NewMailer(cfg),
		loginLimiter: cache.NewLoginFailLimiter(5, 15*time.Minute),
		jwtKeyID:     kid,
	}

	// Pre-load and cache the ECDSA private key at init
	if cfg.JWTPrivateKeyPath != "" {
		keyBytes, err := os.ReadFile(cfg.JWTPrivateKeyPath)
		if err != nil {
			logger.Error().Err(err).Str("path", cfg.JWTPrivateKeyPath).Msg("Failed to read JWT private key — JWT signing will fail")
		} else {
			parsedKey, err := jwt.ParseECPrivateKeyFromPEM(keyBytes)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to parse JWT private key — JWT signing will fail")
			} else {
				s.privateKey = parsedKey
				logger.Info().Msg("JWT private key cached successfully")
			}
		}
	}

	return s
}

// GenerateOTP creates a 6 digit code, sends it via email, and stores pending state in Redis
func (s *Service) GenerateOTP(ctx context.Context, companyName, email, password string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	// Check if already registered
	company, err := s.queries.GetCompanyByEmail(ctx, email)
	if err == nil && company.ID != uuid.Nil {
		return errors.New("a company with this email already exists")
	}

	// Generate 6 digit OTP securely
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return fmt.Errorf("failed to generate secure OTP: %w", err)
	}
	otp := fmt.Sprintf("%06d", n.Int64())
	logger.Info().Str("email", email).Msg("Generated OTP, sending verification email")
	s.mailer.SendAsync(notif.OTPEmail(email, otp))

	// Hash password and OTP
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	hashedOTP, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash OTP: %w", err)
	}

	// Save to Redis
	payload := PendingUserPayload{
		CompanyName:    companyName,
		Email:          email,
		HashedOTP:      string(hashedOTP),
		HashedPassword: string(hashedPassword),
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	redisKey := fmt.Sprintf("pending_user:%s", email)
	if cache.RedisClient == nil {
		return fmt.Errorf("redis not available for OTP storage")
	}
	err = cache.RedisClient.Set(ctx, redisKey, data, 15*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("failed to store pending user in redis: %w", err)
	}

	return nil
}

// VerifyOTP validates the OTP from Redis and creates the user record
func (s *Service) VerifyOTP(ctx context.Context, email, otp string) (*AuthResponse, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	redisKey := fmt.Sprintf("pending_user:%s", email)

	// Fetch from Redis
	if cache.RedisClient == nil {
		return nil, "", errors.New("redis not available — cannot verify OTP")
	}
	data, err := cache.RedisClient.Get(ctx, redisKey).Result()
	if err != nil {
		return nil, "", errors.New("invalid or expired OTP session")
	}

	var payload PendingUserPayload
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return nil, "", errors.New("failed to parse session data")
	}

	// Verify OTP
	logger.Info().Str("email", payload.Email).Msg("Verifying Stateful OTP")
	err = bcrypt.CompareHashAndPassword([]byte(payload.HashedOTP), []byte(otp))
	if err != nil {
		return nil, "", errors.New("incorrect OTP code")
	}

	// Create company in DB directly from Redis data
	company, err := s.queries.CreateCompany(ctx, db.CreateCompanyParams{
		AdminEmail: payload.Email,
		Name:       sql.NullString{String: payload.CompanyName, Valid: true},
		SetupToken: sql.NullString{String: uuid.New().String(), Valid: true},
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create company: %w", err)
	}

	// Update password hash immediately
	err = s.queries.SetCompanyPassword(ctx, db.SetCompanyPasswordParams{
		ID:                company.ID,
		AdminPasswordHash: sql.NullString{String: payload.HashedPassword, Valid: true},
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to set company password: %w", err)
	}

	// Clear the OTP from Redis
	cache.RedisClient.Del(ctx, redisKey)

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
		prefix = utils.GenerateAbbreviation(company.Name.String)
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

	// Brute-force protection: block email after 5 consecutive failures within 15 min
	if blocked, _ := s.loginLimiter.IsBlocked(ctx, email); blocked {
		return nil, "", errors.New("account temporarily locked due to too many failed attempts")
	}

	company, err := s.queries.GetCompanyByEmail(ctx, email)
	if err != nil {
		s.loginLimiter.Increment(ctx, email) //nolint:errcheck
		return nil, "", errors.New("invalid email or password")
	}

	if !company.AdminPasswordHash.Valid {
		return nil, "", errors.New("account not fully set up")
	}

	err = bcrypt.CompareHashAndPassword([]byte(company.AdminPasswordHash.String), []byte(req.Password))
	if err != nil {
		s.loginLimiter.Increment(ctx, email) //nolint:errcheck
		return nil, "", errors.New("invalid email or password")
	}

	// Successful login — clear the failure counter
	s.loginLimiter.Reset(ctx, email)

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

// AdminLogin verifies Super Admin credentials and returns a JWT
func (s *Service) AdminLogin(ctx context.Context, req AdminLoginRequest) (string, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Brute-force protection: same rules as regular login
	if blocked, _ := s.loginLimiter.IsBlocked(ctx, email); blocked {
		return "", errors.New("account temporarily locked due to too many failed attempts")
	}

	admin, err := s.queries.GetSuperAdminByEmail(ctx, email)
	if err != nil {
		s.loginLimiter.Increment(ctx, email) //nolint:errcheck
		return "", errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password))
	if err != nil {
		s.loginLimiter.Increment(ctx, email) //nolint:errcheck
		return "", errors.New("invalid email or password")
	}

	// Successful login — clear the failure counter
	s.loginLimiter.Reset(ctx, email)

	// Generate super admin token (using a dummy UUID for CompanyID since they aren't a company)
	if s.privateKey == nil {
		return "", errors.New("JWT private key is not loaded")
	}

	claims := JWTClaims{
		CompanyID:   uuid.Nil, // Super Admins don't belong to a specific company
		CompanyName: "Super Admin",
		Email:       admin.Email,
		PlanType:    "unlimited",
		AuthStatus:  "active",
		Role:        "super_admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Shorter expiry for admin
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "webtracker-auth",
			Subject:   admin.ID.String(),
			Audience:  jwt.ClaimStrings{"super_admin"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = s.jwtKeyID
	return token.SignedString(s.privateKey)
}

// InitiatePasswordReset starts the password reset flow by sending an OTP.
// Always returns nil regardless of whether the email exists to prevent enumeration.
func (s *Service) InitiatePasswordReset(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	// Check if user exists — but never reveal the result to the caller
	company, err := s.queries.GetCompanyByEmail(ctx, email)
	userExists := err == nil && company.ID != uuid.Nil

	// Generate 6 digit OTP securely
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return fmt.Errorf("failed to generate secure OTP: %w", err)
	}
	otp := fmt.Sprintf("%06d", n.Int64())

	// Only send the email if the account actually exists
	if userExists {
		logger.Info().Str("email", email).Msg("Sending password reset email")
		s.mailer.SendAsync(notif.PasswordResetEmail(email, otp))
	} else {
		logger.Info().Str("email", email).Msg("Password reset requested for non-existent email — returning dummy token")
	}

	hashedOTP, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash OTP: %w", err)
	}

	// Save to Redis
	payload := PendingResetPayload{
		Email:     email,
		HashedOTP: string(hashedOTP),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	redisKey := fmt.Sprintf("pending_reset:%s", email)
	if cache.RedisClient == nil {
		return fmt.Errorf("redis not available for password reset")
	}
	err = cache.RedisClient.Set(ctx, redisKey, data, 15*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("failed to store reset intent in redis: %w", err)
	}

	return nil
}

// CompletePasswordReset verifies the OTP from Redis and updates the password
func (s *Service) CompletePasswordReset(ctx context.Context, req ResetPasswordRequest) error {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	redisKey := fmt.Sprintf("pending_reset:%s", email)

	// Fetch from Redis
	if cache.RedisClient == nil {
		return errors.New("redis not available — cannot verify reset code")
	}
	data, err := cache.RedisClient.Get(ctx, redisKey).Result()
	if err != nil {
		return errors.New("invalid or expired reset session")
	}

	var payload PendingResetPayload
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return errors.New("failed to parse session data")
	}

	// Verify OTP
	otpCode := strings.TrimSpace(req.OTP)
	logger.Info().Int("len", len(otpCode)).Msg("Comparing Reset OTP (Stateful)")
	err = bcrypt.CompareHashAndPassword([]byte(payload.HashedOTP), []byte(otpCode))
	if err != nil {
		return errors.New("incorrect reset code")
	}

	// Get company
	company, err := s.queries.GetCompanyByEmail(ctx, req.Email)
	if err != nil {
		return errors.New("account not found")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	err = s.queries.SetCompanyPassword(ctx, db.SetCompanyPasswordParams{
		ID:                company.ID,
		AdminPasswordHash: sql.NullString{String: string(hashedPassword), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Clear the reset session from Redis
	cache.RedisClient.Del(ctx, redisKey)

	return nil
}

func (s *Service) generateJWT(companyID uuid.UUID, companyName, email, planType, authStatus string) (string, error) {
	if s.privateKey == nil {
		return "", errors.New("JWT private key is not loaded — check JWT_PRIVATE_KEY_PATH")
	}

	role := "authenticated"

	claims := JWTClaims{
		CompanyID:   companyID,
		CompanyName: companyName,
		Email:       email,
		PlanType:    planType,
		AuthStatus:  authStatus,
		Role:        role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "webtracker-auth",
			Subject:   companyID.String(),
			Audience:  jwt.ClaimStrings{role},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = s.jwtKeyID
	return token.SignedString(s.privateKey)
}
