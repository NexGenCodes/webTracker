package usecase

import (
	"context"
	"database/sql"
	"fmt"
	
	"webtracker-bot/internal/adapter/db"
)

// ConfigUsecase handles system configuration, user preferences, and bot core authorization.
type ConfigUsecase struct {
	repo db.Querier
	pool *sql.DB
}

func NewConfigUsecase(repo db.Querier, pool *sql.DB) *ConfigUsecase {
	return &ConfigUsecase{
		repo: repo,
		pool: pool,
	}
}

// Ping checks if the database is responding.
func (u *ConfigUsecase) Ping(ctx context.Context) error {
	if u.pool == nil {
		return fmt.Errorf("database pool is not initialized")
	}
	return u.pool.PingContext(ctx)
}

// GetSystemConfig retrieves a system config value. Returns empty string if not found.
func (u *ConfigUsecase) GetSystemConfig(ctx context.Context, key string) (string, error) {
	val, err := u.repo.GetSystemConfig(ctx, key)
	if err == sql.ErrNoRows {
		return "", nil // default handling
	}
	if err != nil {
		return "", fmt.Errorf("failed to get system config: %w", err)
	}
	// val is of type string (from pgx)
	return val, nil
}

// SetSystemConfig saves a system config value.
func (u *ConfigUsecase) SetSystemConfig(ctx context.Context, key, value string) error {
	params := db.SetSystemConfigParams{
		Key:   key,
		Value: value,
	}
	err := u.repo.SetSystemConfig(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to set system config: %w", err)
	}
	return nil
}

// GetUserLanguage retrieves language preference strictly.
func (u *ConfigUsecase) GetUserLanguage(ctx context.Context, jid string) (string, error) {
	lang, err := u.repo.GetUserLanguage(ctx, jid)
	if err == sql.ErrNoRows {
		return "en", nil
	}
	if err != nil {
		return "en", fmt.Errorf("failed to get user language: %w", err)
	}
	return lang, nil
}

// SetUserLanguage stores the user's preferred language.
func (u *ConfigUsecase) SetUserLanguage(ctx context.Context, jid, lang string) error {
	params := db.SetUserLanguageParams{
		Jid:      jid,
		Language: lang,
	}
	err := u.repo.SetUserLanguage(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to set user language: %w", err)
	}
	return nil
}

// GetGroupAuthority checks bot authorization caching. Returns (isAuthorized, exists, error).
func (u *ConfigUsecase) GetGroupAuthority(ctx context.Context, jid string) (bool, bool, error) {
	row, err := u.repo.GetGroupAuthority(ctx, jid)
	if err == sql.ErrNoRows {
		return false, false, nil
	}
	if err != nil {
		return false, false, fmt.Errorf("failed to get group authority: %w", err)
	}
	return row.IsAuthorized, true, nil
}

// SetGroupAuthority caches the bot's group admin state.
func (u *ConfigUsecase) SetGroupAuthority(ctx context.Context, jid string, isAuthorized bool) error {
	params := db.SetGroupAuthorityParams{
		Jid:          jid,
		IsAuthorized: isAuthorized,
	}
	err := u.repo.SetGroupAuthority(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to set group authority: %w", err)
	}
	return nil
}

// HasAuthorizedGroups strictly checks if the bot is admin anywhere.
func (u *ConfigUsecase) HasAuthorizedGroups(ctx context.Context) (bool, error) {
	count, err := u.repo.HasAuthorizedGroups(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check authorized groups: %w", err)
	}
	return count > 0, nil
}

// GetAuthorizedGroups returns all group JIDs where the bot has authority.
func (u *ConfigUsecase) GetAuthorizedGroups(ctx context.Context) ([]string, error) {
	groups, err := u.repo.GetAuthorizedGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list authorized groups: %w", err)
	}
	return groups, nil
}

// CountAuthorizedGroups returns the total number of authorized groups.
func (u *ConfigUsecase) CountAuthorizedGroups(ctx context.Context) (int64, error) {
	count, err := u.repo.CountAuthorizedGroups(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count authorized groups: %w", err)
	}
	return count, nil
}
