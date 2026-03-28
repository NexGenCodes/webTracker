package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"webtracker-bot/internal/adapter/db"
	"webtracker-bot/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockQuerier is a manually implementation of db.Querier using testify/mock
type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) GetShipment(ctx context.Context, id string) (db.Shipment, error) {
	args := m.Called(ctx, id)
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

func (m *MockQuerier) TransitionStatusToIntransit(ctx context.Context, now sql.NullTime) ([]db.TransitionStatusToIntransitRow, error) {
	args := m.Called(ctx, now)
	return args.Get(0).([]db.TransitionStatusToIntransitRow), args.Error(1)
}

func (m *MockQuerier) TransitionStatusToOutForDelivery(ctx context.Context, now sql.NullTime) ([]db.TransitionStatusToOutForDeliveryRow, error) {
	args := m.Called(ctx, now)
	return args.Get(0).([]db.TransitionStatusToOutForDeliveryRow), args.Error(1)
}

func (m *MockQuerier) TransitionStatusToDelivered(ctx context.Context, now sql.NullTime) ([]db.TransitionStatusToDeliveredRow, error) {
	args := m.Called(ctx, now)
	return args.Get(0).([]db.TransitionStatusToDeliveredRow), args.Error(1)
}

// Implement remaining methods with no-ops or panics if not used in tests
func (m *MockQuerier) BulkUpdateStatus(ctx context.Context, arg db.BulkUpdateStatusParams) error { return nil }
func (m *MockQuerier) CountAuthorizedGroups(ctx context.Context) (int64, error) { return 0, nil }
func (m *MockQuerier) CountCreatedSince(ctx context.Context, createdAt sql.NullTime) (int64, error) { return 0, nil }
func (m *MockQuerier) CountDeliveredSince(ctx context.Context, updatedAt sql.NullTime) (int64, error) { return 0, nil }
func (m *MockQuerier) CountShipments(ctx context.Context) (int64, error) { return 0, nil }
func (m *MockQuerier) CountShipmentsByStatus(ctx context.Context) (db.CountShipmentsByStatusRow, error) { return db.CountShipmentsByStatusRow{}, nil }
func (m *MockQuerier) DeleteDeliveredShipments(ctx context.Context) error { return nil }
func (m *MockQuerier) DeleteShipment(ctx context.Context, id string) error { return nil }
func (m *MockQuerier) FindSimilarShipment(ctx context.Context, arg db.FindSimilarShipmentParams) (string, error) { return "", nil }
func (m *MockQuerier) GetAuthorizedGroups(ctx context.Context) ([]string, error) { return nil, nil }
func (m *MockQuerier) GetGroupAuthority(ctx context.Context, jid string) (db.GetGroupAuthorityRow, error) { 
	args := m.Called(ctx, jid)
	return args.Get(0).(db.GetGroupAuthorityRow), args.Error(1)
}
func (m *MockQuerier) GetLastShipmentIDForUser(ctx context.Context, userJid string) (string, error) { return "", nil }
func (m *MockQuerier) GetRecentEvents(ctx context.Context, limit int32) ([]db.Telemetry, error) { return nil, nil }
func (m *MockQuerier) GetSystemConfig(ctx context.Context, key string) (string, error) { 
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}
func (m *MockQuerier) GetTelemetryStats(ctx context.Context, createdAt sql.NullTime) ([]db.GetTelemetryStatsRow, error) { return nil, nil }
func (m *MockQuerier) GetUserLanguage(ctx context.Context, jid string) (string, error) { return "en", nil }
func (m *MockQuerier) HasAuthorizedGroups(ctx context.Context) (int64, error) { return 0, nil }
func (m *MockQuerier) ListAllShipments(ctx context.Context) ([]db.Shipment, error) { return nil, nil }
func (m *MockQuerier) ListShipments(ctx context.Context, arg db.ListShipmentsParams) ([]db.Shipment, error) { return nil, nil }
func (m *MockQuerier) RecordEvent(ctx context.Context, arg db.RecordEventParams) error { return nil }
func (m *MockQuerier) RunAgedCleanup(ctx context.Context, arg db.RunAgedCleanupParams) error { return nil }
func (m *MockQuerier) SetGroupAuthority(ctx context.Context, arg db.SetGroupAuthorityParams) error { 
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockQuerier) SetSystemConfig(ctx context.Context, arg db.SetSystemConfigParams) error { return nil }
func (m *MockQuerier) SetUserLanguage(ctx context.Context, arg db.SetUserLanguageParams) error { return nil }
func (m *MockQuerier) UpdateShipmentFieldCargoType(ctx context.Context, arg db.UpdateShipmentFieldCargoTypeParams) error { return nil }
func (m *MockQuerier) UpdateShipmentFieldDestination(ctx context.Context, arg db.UpdateShipmentFieldDestinationParams) error { return nil }
func (m *MockQuerier) UpdateShipmentFieldExpectedDeliveryTime(ctx context.Context, arg db.UpdateShipmentFieldExpectedDeliveryTimeParams) error { return nil }
func (m *MockQuerier) UpdateShipmentFieldOrigin(ctx context.Context, arg db.UpdateShipmentFieldOriginParams) error { return nil }
func (m *MockQuerier) UpdateShipmentFieldOutfordeliveryTime(ctx context.Context, arg db.UpdateShipmentFieldOutfordeliveryTimeParams) error { return nil }
func (m *MockQuerier) UpdateShipmentFieldRecipientAddress(ctx context.Context, arg db.UpdateShipmentFieldRecipientAddressParams) error { return nil }
func (m *MockQuerier) UpdateShipmentFieldRecipientEmail(ctx context.Context, arg db.UpdateShipmentFieldRecipientEmailParams) error { return nil }
func (m *MockQuerier) UpdateShipmentFieldRecipientID(ctx context.Context, arg db.UpdateShipmentFieldRecipientIDParams) error { return nil }
func (m *MockQuerier) UpdateShipmentFieldRecipientName(ctx context.Context, arg db.UpdateShipmentFieldRecipientNameParams) error { return nil }
func (m *MockQuerier) UpdateShipmentFieldRecipientPhone(ctx context.Context, arg db.UpdateShipmentFieldRecipientPhoneParams) error { return nil }
func (m *MockQuerier) UpdateShipmentFieldScheduledTransitTime(ctx context.Context, arg db.UpdateShipmentFieldScheduledTransitTimeParams) error { return nil }
func (m *MockQuerier) UpdateShipmentFieldSenderName(ctx context.Context, arg db.UpdateShipmentFieldSenderNameParams) error { return nil }
func (m *MockQuerier) UpdateShipmentFieldSenderPhone(ctx context.Context, arg db.UpdateShipmentFieldSenderPhoneParams) error { return nil }


func TestShipmentUsecase_Deep(t *testing.T) {
	ctx := context.Background()
	repo := new(MockQuerier)
	uc := usecase.NewShipmentUsecase(repo, nil)

	t.Run("Track_Success", func(t *testing.T) {
		repo.On("GetShipment", ctx, "AWB-101").Return(db.Shipment{TrackingID: "AWB-101", RecipientName: sql.NullString{String: "John", Valid: true}}, nil).Once()
		res, err := uc.Track(ctx, "AWB-101")
		assert.NoError(t, err)
		assert.Equal(t, "John", res.RecipientName.String)
		repo.AssertExpectations(t)
	})

	t.Run("Create_ValidationFailure", func(t *testing.T) {
		err := uc.Create(ctx, db.CreateShipmentParams{TrackingID: ""})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tracking ID is required")
	})

	t.Run("ProcessTransitions_DeepFlow", func(t *testing.T) {
		now := time.Now()
		nullNow := sql.NullTime{Time: now, Valid: true}
		
		// 1. Pending -> InTransit
		repo.On("TransitionStatusToIntransit", ctx, nullNow).Return([]db.TransitionStatusToIntransitRow{
			{TrackingID: "T1", NewStatus: sql.NullString{String: "intransit", Valid: true}, UserJid: "U1"},
		}, nil).Once()
		
		// 2. InTransit -> OutForDelivery
		repo.On("TransitionStatusToOutForDelivery", ctx, nullNow).Return([]db.TransitionStatusToOutForDeliveryRow{
			{TrackingID: "T2", NewStatus: sql.NullString{String: "outfordelivery", Valid: true}, UserJid: "U2"},
		}, nil).Once()

		// 3. OutForDelivery -> Delivered
		repo.On("TransitionStatusToDelivered", ctx, nullNow).Return([]db.TransitionStatusToDeliveredRow{
			{TrackingID: "T3", NewStatus: sql.NullString{String: "delivered", Valid: true}, UserJid: "U3"},
		}, nil).Once()

		results, err := uc.ProcessTransitions(ctx, now)
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
	uc := usecase.NewConfigUsecase(repo, nil)

	t.Run("RBAC_AuthorityCheck", func(t *testing.T) {
		repo.On("GetGroupAuthority", ctx, "group1").Return(db.GetGroupAuthorityRow{IsAuthorized: true}, nil).Once()
		isAuth, exists, err := uc.GetGroupAuthority(ctx, "group1")
		assert.NoError(t, err)
		assert.True(t, isAuth)
		assert.True(t, exists)
		repo.AssertExpectations(t)
	})

	t.Run("SystemConfig_Fallback", func(t *testing.T) {
		repo.On("GetSystemConfig", ctx, "missing_key").Return("", sql.ErrNoRows).Once()
		val, err := uc.GetSystemConfig(ctx, "missing_key")
		assert.NoError(t, err)
		assert.Equal(t, "", val)
		repo.AssertExpectations(t)
	})
}
