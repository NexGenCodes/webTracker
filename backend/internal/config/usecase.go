package config

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"webtracker-bot/internal/database/db"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type Usecase struct {
	repo db.Querier
	pool *sql.DB
}

func NewUsecase(repo db.Querier, pool *sql.DB) *Usecase {
	return &Usecase{repo: repo, pool: pool}
}

func (u *Usecase) Ping(ctx context.Context) error {
	if u.pool == nil {
		return fmt.Errorf("database pool is not initialized")
	}
	return u.pool.PingContext(ctx)
}

func (u *Usecase) GetAllCompanies(ctx context.Context) ([]uuid.UUID, error) {
	return u.repo.GetAllCompanies(ctx)
}

func (u *Usecase) GetAllActiveCompanies(ctx context.Context) ([]db.Company, error) {
	return u.repo.GetAllActiveCompanies(ctx)
}

func (u *Usecase) GetSystemConfig(ctx context.Context, companyID uuid.UUID, key string) (string, error) {
	val, err := u.repo.GetSystemConfig(ctx, db.GetSystemConfigParams{CompanyID: companyID, Key: key})
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get system config: %w", err)
	}
	return val, nil
}

func (u *Usecase) SetSystemConfig(ctx context.Context, companyID uuid.UUID, key, value string) error {
	return u.repo.SetSystemConfig(ctx, db.SetSystemConfigParams{CompanyID: companyID, Key: key, Value: value})
}

func (u *Usecase) GetUserLanguage(ctx context.Context, companyID uuid.UUID, jid string) (string, error) {
	lang, err := u.repo.GetUserLanguage(ctx, db.GetUserLanguageParams{CompanyID: companyID, Jid: jid})
	if err == sql.ErrNoRows {
		return "en", nil
	}
	if err != nil {
		return "en", err
	}
	return lang, nil
}

func (u *Usecase) SetUserLanguage(ctx context.Context, companyID uuid.UUID, jid, lang string) error {
	return u.repo.SetUserLanguage(ctx, db.SetUserLanguageParams{CompanyID: companyID, Jid: jid, Language: lang})
}

func (u *Usecase) GetGroupAuthority(ctx context.Context, companyID uuid.UUID, jid string) (bool, bool, error) {
	row, err := u.repo.GetGroupAuthority(ctx, db.GetGroupAuthorityParams{CompanyID: companyID, Jid: jid})
	if err == sql.ErrNoRows {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	return row.IsAuthorized, true, nil
}

func (u *Usecase) SetGroupAuthority(ctx context.Context, companyID uuid.UUID, jid string, isAuthorized bool) error {
	return u.repo.SetGroupAuthority(ctx, db.SetGroupAuthorityParams{CompanyID: companyID, Jid: jid, IsAuthorized: isAuthorized})
}

func (u *Usecase) HasAuthorizedGroups(ctx context.Context, companyID uuid.UUID) (bool, error) {
	count, err := u.repo.HasAuthorizedGroups(ctx, companyID)
	return count > 0, err
}

func (u *Usecase) GetAuthorizedGroups(ctx context.Context, companyID uuid.UUID) ([]string, error) {
	return u.repo.GetAuthorizedGroups(ctx, companyID)
}

func (u *Usecase) CountAuthorizedGroups(ctx context.Context, companyID uuid.UUID) (int64, error) {
	return u.repo.CountAuthorizedGroups(ctx, companyID)
}

func (u *Usecase) GetCompanyByID(ctx context.Context, companyID uuid.UUID) (db.Company, error) {
	return u.repo.GetCompanyByID(ctx, companyID)
}

func (u *Usecase) UpdateCompanySettings(ctx context.Context, companyID uuid.UUID, name, adminEmail, logoUrl string) error {
	return u.repo.UpdateCompanySettings(ctx, db.UpdateCompanySettingsParams{
		ID:         companyID,
		Name:       sql.NullString{String: name, Valid: name != ""},
		AdminEmail: adminEmail,
		LogoUrl:    sql.NullString{String: logoUrl, Valid: true},
	})
}

func (u *Usecase) UpdateCompanyAuthStatus(ctx context.Context, companyID uuid.UUID, authStatus string) error {
	return u.repo.UpdateCompanyAuthStatus(ctx, db.UpdateCompanyAuthStatusParams{
		ID:         companyID,
		AuthStatus: sql.NullString{String: authStatus, Valid: true},
	})
}

func (u *Usecase) UpdateCompanySubscriptionStatus(ctx context.Context, companyID uuid.UUID, subStatus, planType string) error {
	return u.repo.UpdateCompanySubscriptionWithPlan(ctx, db.UpdateCompanySubscriptionWithPlanParams{
		ID:                 companyID,
		SubscriptionStatus: sql.NullString{String: subStatus, Valid: subStatus != ""},
		PlanType:           planType,
	})
}

func (u *Usecase) CreateCompany(ctx context.Context, name, adminEmail, setupToken string) (db.Company, error) {
	return u.repo.CreateCompany(ctx, db.CreateCompanyParams{
		Name:       sql.NullString{String: name, Valid: name != ""},
		AdminEmail: adminEmail,
		SetupToken: sql.NullString{String: setupToken, Valid: true},
	})
}

// RecordPayment returns the new payment ID if successful, or 0 if it was a duplicate
func (u *Usecase) RecordPayment(ctx context.Context, companyID uuid.UUID, reference string, amount float64, status string) (int32, error) {
	id, err := u.repo.RecordPayment(ctx, db.RecordPaymentParams{
		CompanyID: uuid.NullUUID{UUID: companyID, Valid: true},
		Reference: reference,
		Amount:    sql.NullFloat64{Float64: amount, Valid: true},
		Status:    sql.NullString{String: status, Valid: true},
	})
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}

func (u *Usecase) UpdateCompanyWhatsAppPhone(ctx context.Context, companyID uuid.UUID, phone string) error {
	return u.repo.UpdateCompanyWhatsAppPhone(ctx, db.UpdateCompanyWhatsAppPhoneParams{
		WhatsappPhone: sql.NullString{String: phone, Valid: phone != ""},
		ID:            companyID,
	})
}

func (u *Usecase) DeleteCompany(ctx context.Context, companyID uuid.UUID) error {
	return u.repo.DeleteCompany(ctx, companyID)
}

func (uc *Usecase) GetActivePlans(ctx context.Context) ([]db.GetActivePlansRow, error) {
	return uc.repo.GetActivePlans(ctx)
}

func (uc *Usecase) GetPlanByID(ctx context.Context, id string) (db.GetPlanByIDRow, error) {
	return uc.repo.GetPlanByID(ctx, id)
}

func (u *Usecase) LogAudit(ctx context.Context, actorEmail, action string, targetCompanyID uuid.UUID, details map[string]interface{}) error {
	var detailsRaw pqtype.NullRawMessage
	if details != nil {
		data, err := json.Marshal(details)
		if err == nil {
			detailsRaw = pqtype.NullRawMessage{RawMessage: data, Valid: true}
		}
	}

	return u.repo.LogAudit(ctx, db.LogAuditParams{
		ActorEmail:      actorEmail,
		Action:          action,
		TargetCompanyID: uuid.NullUUID{UUID: targetCompanyID, Valid: targetCompanyID != uuid.Nil},
		Details:         detailsRaw,
	})
}

func (u *Usecase) GetAuditLogs(ctx context.Context, limit, offset int32) ([]db.AuditLog, error) {
	return u.repo.GetAuditLogs(ctx, db.GetAuditLogsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (u *Usecase) GetPlatformAnalytics(ctx context.Context) (db.GetPlatformAnalyticsRow, error) {
	return u.repo.GetPlatformAnalytics(ctx)
}

func (u *Usecase) UpdateCompanyPlan(ctx context.Context, companyID uuid.UUID, planType string) error {
	return u.repo.UpdateCompanyPlan(ctx, db.UpdateCompanyPlanParams{
		ID:       companyID,
		PlanType: sql.NullString{String: planType, Valid: true},
	})
}

func (u *Usecase) UpdateCompanySubscription(ctx context.Context, companyID uuid.UUID, subStatus string, expiry time.Time) error {
	return u.repo.UpdateCompanySubscription(ctx, db.UpdateCompanySubscriptionParams{
		ID:                 companyID,
		SubscriptionStatus: sql.NullString{String: subStatus, Valid: true},
		SubscriptionExpiry: sql.NullTime{Time: expiry, Valid: !expiry.IsZero()},
	})
}

func (u *Usecase) GetCompanyPayments(ctx context.Context, companyID uuid.UUID, limit, offset int32) ([]db.Payment, error) {
	return u.repo.GetCompanyPayments(ctx, db.GetCompanyPaymentsParams{
		CompanyID: uuid.NullUUID{UUID: companyID, Valid: true},
		Limit:     limit,
		Offset:    offset,
	})
}

