package config

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"webtracker-bot/internal/database/db"
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

func (u *Usecase) GetCompanyBySetupToken(ctx context.Context, setupToken string) (db.Company, error) {
	return u.repo.GetCompanyBySetupToken(ctx, sql.NullString{String: setupToken, Valid: true})
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

func (u *Usecase) UpdateCompanySubscriptionStatus(ctx context.Context, companyID uuid.UUID, subStatus string) error {
	return u.repo.UpdateCompanySubscriptionStatus(ctx, db.UpdateCompanySubscriptionStatusParams{
		ID:                 companyID,
		SubscriptionStatus: sql.NullString{String: subStatus, Valid: true},
	})
}

func (u *Usecase) CreateCompany(ctx context.Context, name, adminEmail, setupToken string) (db.Company, error) {
	return u.repo.CreateCompany(ctx, db.CreateCompanyParams{
		Name:       sql.NullString{String: name, Valid: name != ""},
		AdminEmail: adminEmail,
		SetupToken: sql.NullString{String: setupToken, Valid: true},
	})
}

func (u *Usecase) RegenerateSetupToken(ctx context.Context, companyID uuid.UUID, newToken string) error {
	return u.repo.RegenerateSetupToken(ctx, db.RegenerateSetupTokenParams{
		ID:         companyID,
		SetupToken: sql.NullString{String: newToken, Valid: true},
	})
}

// UpdateCompanyWhatsAppPhone updates the WhatsApp phone for a company.
// Uses direct SQL to avoid SQLC regeneration requirement.
func (u *Usecase) UpdateCompanyWhatsAppPhone(ctx context.Context, companyID uuid.UUID, phone string) error {
	_, err := u.pool.ExecContext(ctx,
		"UPDATE companies SET whatsapp_phone = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		phone, companyID,
	)
	return err
}


