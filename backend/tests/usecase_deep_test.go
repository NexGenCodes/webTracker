package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"webtracker-bot/internal/database/db"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/config"
	)

// MockQuerier is a manually implementation of db.Querier using testify/mock
type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) GetShipment(ctx context.Context, arg db.GetShipmentParams) (db.Shipment, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Shipment), args.Error(1)
}

func (m *MockQuerier) CreateShipment(ctx context.Context, arg db.CreateShipmentParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerier) UpdateShipmentStatus(ctx context.Context, arg db.UpdateShipmentStatusParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerier) TransitionStatusToIntransit(ctx context.Context, arg db.TransitionStatusToIntransitParams) ([]db.TransitionStatusToIntransitRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.TransitionStatusToIntransitRow), args.Error(1)
}

func (m *MockQuerier) TransitionStatusToOutForDelivery(ctx context.Context, arg db.TransitionStatusToOutForDeliveryParams) ([]db.TransitionStatusToOutForDeliveryRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.TransitionStatusToOutForDeliveryRow), args.Error(1)
}

func (m *MockQuerier) TransitionStatusToDelivered(ctx context.Context, arg db.TransitionStatusToDeliveredParams) ([]db.TransitionStatusToDeliveredRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.TransitionStatusToDeliveredRow), args.Error(1)
}

// Implement remaining methods with no-ops
func (m *MockQuerier) BulkUpdateStatus(ctx context.Context, arg db.BulkUpdateStatusParams) error {
	return nil
}
func (m *MockQuerier) BulkDeleteShipments(ctx context.Context, arg db.BulkDeleteShipmentsParams) (sql.Result, error) {
	return mockResult{}, nil
}
func (m *MockQuerier) CountAuthorizedGroups(ctx context.Context, companyID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) CountCreatedSince(ctx context.Context, arg db.CountCreatedSinceParams) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) CountDeliveredSince(ctx context.Context, arg db.CountDeliveredSinceParams) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) CountShipments(ctx context.Context, companyID uuid.NullUUID) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) CountShipmentsByStatus(ctx context.Context, companyID uuid.NullUUID) (db.CountShipmentsByStatusRow, error) {
	return db.CountShipmentsByStatusRow{}, nil
}
func (m *MockQuerier) DeleteDeliveredShipments(ctx context.Context, companyID uuid.NullUUID) error {
	return nil
}
func (m *MockQuerier) DeleteShipment(ctx context.Context, arg db.DeleteShipmentParams) error {
	return nil
}
func (m *MockQuerier) FindSimilarShipment(ctx context.Context, arg db.FindSimilarShipmentParams) (string, error) {
	return "", nil
}
func (m *MockQuerier) GetAllCompanies(ctx context.Context) ([]uuid.UUID, error) { return nil, nil }
func (m *MockQuerier) GetAuthorizedGroups(ctx context.Context, companyID uuid.UUID) ([]string, error) {
	return nil, nil
}
func (m *MockQuerier) GetCompanyByID(ctx context.Context, id uuid.UUID) (db.Company, error) {
	return db.Company{}, nil
}
func (m *MockQuerier) GetGroupAuthority(ctx context.Context, arg db.GetGroupAuthorityParams) (db.GetGroupAuthorityRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.GetGroupAuthorityRow), args.Error(1)
}
func (m *MockQuerier) GetLastShipmentIDForUser(ctx context.Context, arg db.GetLastShipmentIDForUserParams) (string, error) {
	return "", nil
}
func (m *MockQuerier) GetRecentEvents(ctx context.Context, arg db.GetRecentEventsParams) ([]db.Telemetry, error) {
	return nil, nil
}
func (m *MockQuerier) GetSystemConfig(ctx context.Context, arg db.GetSystemConfigParams) (string, error) {
	args := m.Called(ctx, arg)
	return args.String(0), args.Error(1)
}
func (m *MockQuerier) GetTelemetryStats(ctx context.Context, arg db.GetTelemetryStatsParams) ([]db.GetTelemetryStatsRow, error) {
	return nil, nil
}
func (m *MockQuerier) GetUserLanguage(ctx context.Context, arg db.GetUserLanguageParams) (string, error) {
	return "en", nil
}
func (m *MockQuerier) HasAuthorizedGroups(ctx context.Context, companyID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) ListAllShipments(ctx context.Context, companyID uuid.NullUUID) ([]db.Shipment, error) {
	return nil, nil
}
func (m *MockQuerier) ListShipments(ctx context.Context, arg db.ListShipmentsParams) ([]db.Shipment, error) {
	return nil, nil
}
func (m *MockQuerier) RecordEvent(ctx context.Context, arg db.RecordEventParams) error { return nil }
func (m *MockQuerier) RunAgedCleanup(ctx context.Context, arg db.RunAgedCleanupParams) (sql.Result, error) {
	return mockResult{}, nil
}
func (m *MockQuerier) SetGroupAuthority(ctx context.Context, arg db.SetGroupAuthorityParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockQuerier) SetSystemConfig(ctx context.Context, arg db.SetSystemConfigParams) error {
	return nil
}
func (m *MockQuerier) SetUserLanguage(ctx context.Context, arg db.SetUserLanguageParams) error {
	return nil
}
func (m *MockQuerier) UpdateShipmentDynamic(ctx context.Context, arg db.UpdateShipmentDynamicParams) error {
	return nil
}
func (m *MockQuerier) CreateCompany(ctx context.Context, arg db.CreateCompanyParams) (db.Company, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Company), args.Error(1)
}
func (m *MockQuerier) GetCompanyByEmail(ctx context.Context, adminEmail string) (db.Company, error) {
	args := m.Called(ctx, adminEmail)
	return args.Get(0).(db.Company), args.Error(1)
}
func (m *MockQuerier) GetAllActiveCompanies(ctx context.Context) ([]db.Company, error) {
	return nil, nil
}

func (m *MockQuerier) GetActivePlans(ctx context.Context) ([]db.GetActivePlansRow, error) {
	return nil, nil
}

func (m *MockQuerier) GetPlanByID(ctx context.Context, id string) (db.GetPlanByIDRow, error) {
	return db.GetPlanByIDRow{}, nil
}

func (m *MockQuerier) UpdatePlanPrice(ctx context.Context, arg db.UpdatePlanPriceParams) error {
	return nil
}

func (m *MockQuerier) UpdateCompanyAuthStatus(ctx context.Context, arg db.UpdateCompanyAuthStatusParams) error {
	return nil
}
func (m *MockQuerier) UpdateCompanySettings(ctx context.Context, arg db.UpdateCompanySettingsParams) error {
	return nil
}
func (m *MockQuerier) UpdateCompanySubscriptionStatus(ctx context.Context, arg db.UpdateCompanySubscriptionStatusParams) error {
	return nil
}
func (m *MockQuerier) SetCompanyPassword(ctx context.Context, arg db.SetCompanyPasswordParams) error {
	return nil
}
func (m *MockQuerier) UpdateCompanyOnboarding(ctx context.Context, arg db.UpdateCompanyOnboardingParams) error {
	return nil
}
func (m *MockQuerier) RecordPayment(ctx context.Context, arg db.RecordPaymentParams) (int32, error) {
	return 1, nil
}

// mockResult implements sql.Result for mock returns
type mockResult struct{}

func (mockResult) LastInsertId() (int64, error) { return 0, nil }
func (mockResult) RowsAffected() (int64, error) { return 0, nil }

// Test Company ID for all tests
var testCompanyID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

func TestShipmentUsecase_Deep(t *testing.T) {
	ctx := context.Background()
	repo := new(MockQuerier)
	uc := shipment.NewUsecase(repo, nil)

	t.Run("Track_Success", func(t *testing.T) {
		getParams := db.GetShipmentParams{
			CompanyID:  uuid.NullUUID{UUID: testCompanyID, Valid: true},
			TrackingID: "AWB-101",
		}
		repo.On("GetShipment", ctx, getParams).Return(db.Shipment{TrackingID: "AWB-101", RecipientName: sql.NullString{String: "John", Valid: true}}, nil).Once()
		res, err := uc.Track(ctx, testCompanyID, "AWB-101")
		assert.NoError(t, err)
		assert.Equal(t, "John", res.RecipientName.String)
		repo.AssertExpectations(t)
	})

	t.Run("Create_ValidationFailure", func(t *testing.T) {
		err := uc.Create(ctx, testCompanyID, db.CreateShipmentParams{TrackingID: ""})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tracking ID is required")
	})

	t.Run("ProcessTransitions_DeepFlow", func(t *testing.T) {
		now := time.Now()
		nullNow := sql.NullTime{Time: now, Valid: true}
		companyNullUUID := uuid.NullUUID{UUID: testCompanyID, Valid: true}

		// 1. Pending -> InTransit
		intransitParams := db.TransitionStatusToIntransitParams{CompanyID: companyNullUUID, ScheduledTransitTime: nullNow}
		repo.On("TransitionStatusToIntransit", ctx, intransitParams).Return([]db.TransitionStatusToIntransitRow{
			{TrackingID: "T1", NewStatus: sql.NullString{String: "intransit", Valid: true}, UserJid: "U1"},
		}, nil).Once()

		// 2. InTransit -> OutForDelivery
		ofdParams := db.TransitionStatusToOutForDeliveryParams{CompanyID: companyNullUUID, OutfordeliveryTime: nullNow}
		repo.On("TransitionStatusToOutForDelivery", ctx, ofdParams).Return([]db.TransitionStatusToOutForDeliveryRow{
			{TrackingID: "T2", NewStatus: sql.NullString{String: "outfordelivery", Valid: true}, UserJid: "U2"},
		}, nil).Once()

		// 3. OutForDelivery -> Delivered
		deliveredParams := db.TransitionStatusToDeliveredParams{CompanyID: companyNullUUID, ExpectedDeliveryTime: nullNow}
		repo.On("TransitionStatusToDelivered", ctx, deliveredParams).Return([]db.TransitionStatusToDeliveredRow{
			{TrackingID: "T3", NewStatus: sql.NullString{String: "delivered", Valid: true}, UserJid: "U3"},
		}, nil).Once()

		results, err := uc.ProcessTransitions(ctx, testCompanyID, now)
		assert.NoError(t, err)
		assert.Len(t, results, 3)
		assert.Equal(t, "T1", results[0].TrackingID)
		assert.Equal(t, "intransit", results[0].NewStatus)
		repo.AssertExpectations(t)
	})
}

func TestConfigUsecase_Deep(t *testing.T) {
	ctx := context.Background()
	repo := new(MockQuerier)
	uc := config.NewUsecase(repo, nil)

	t.Run("RBAC_AuthorityCheck", func(t *testing.T) {
		authParams := db.GetGroupAuthorityParams{CompanyID: testCompanyID, Jid: "group1"}
		repo.On("GetGroupAuthority", ctx, authParams).Return(db.GetGroupAuthorityRow{IsAuthorized: true}, nil).Once()
		isAuth, exists, err := uc.GetGroupAuthority(ctx, testCompanyID, "group1")
		assert.NoError(t, err)
		assert.True(t, isAuth)
		assert.True(t, exists)
		repo.AssertExpectations(t)
	})

	t.Run("SystemConfig_Fallback", func(t *testing.T) {
		cfgParams := db.GetSystemConfigParams{CompanyID: testCompanyID, Key: "missing_key"}
		repo.On("GetSystemConfig", ctx, cfgParams).Return("", sql.ErrNoRows).Once()
		val, err := uc.GetSystemConfig(ctx, testCompanyID, "missing_key")
		assert.NoError(t, err)
		assert.Equal(t, "", val)
		repo.AssertExpectations(t)
	})
}


