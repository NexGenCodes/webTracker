package commands

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"
	"webtracker-bot/internal/notif"
	"webtracker-bot/internal/parser"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/utils"
)

// EditHandler handles !edit [trackingID] [field] [value] or !edit [field] [value]
type EditHandler struct {
	CompanyName   string
	CompanyPrefix string
	Sender        models.WhatsAppSender
	Cfg           *config.Config
}

func (h *EditHandler) Execute(ctx context.Context, shipUC models.ShipmentUsecase, configUC models.ConfigUsecase, companyID uuid.UUID, args []string, lang string, isAdmin bool) Result {
	if len(args) < 1 {
		return Result{Message: "✏️ *EDIT SHIPMENT*\n\nUsage: `!edit [TrackingID] [Updates...]` or `!edit [Updates...]` (targets last shipment)\n\n*Example:* `!edit LGS-1234 name: John, departure: tomorrow`"}
	}

	jid := utils.GetJID(ctx)
	var trackingID string
	var startIdx int

	// 1. Identify Target Shipment
	// Pattern: 3 uppercase letters, hyphen, 4 digits
	idPattern := regexp.MustCompile(`^[A-Z]{3}-\d{4,9}$`)
	if idPattern.MatchString(args[0]) {
		trackingID = args[0]
		startIdx = 1
	} else {
		// Contextual Lookup: Fetch last shipment for this user
		var err error
		trackingID, err = shipUC.GetLastForUser(ctx, companyID, jid)
		if err != nil || trackingID == "" {
			return Result{Message: i18n.T(i18nLang(lang), "ERR_CONTEXT_ERROR")}
		}
		startIdx = 0
	}

	// 2. Parse Updates
	updateText := strings.Join(args[startIdx:], " ")
	updates := parser.ParseEditPairs(updateText)

	// Fallback for single field (e.g., !edit name Mark) if parser didn't find clear anchors
	if len(updates) == 0 && len(args[startIdx:]) >= 2 {
		// This handles the old style: !edit field value
		field := args[startIdx]
		value := strings.Join(args[startIdx+1:], " ")
		// Reuse canonical mapping logic
		normField := strings.ToLower(field)
		dbField := ""
		switch normField {
		case "receiver", "name", "recipient", "reciever":
			dbField = "recipient_name"
		case "phone", "mobile", "number", "contact":
			dbField = "recipient_phone"
		case "address", "addr":
			dbField = "recipient_address"
		case "destination", "country", "to":
			dbField = "destination"
		case "origin", "from", "source":
			dbField = "origin"
		case "id", "passport", "identification":
			dbField = "recipient_id"
		case "email", "mail":
			dbField = "recipient_email"
		case "departure", "transit":
			dbField = "scheduled_transit_time"
		case "arrival", "delivery":
			dbField = "expected_delivery_time"
		case "outfordelivery", "out_for_delivery":
			dbField = "outfordelivery_time"
		case "cargo", "type", "content":
			dbField = "cargo_type"
		case "weight":
			dbField = "weight"
		}
		if dbField != "" {
			updates[dbField] = value
		}
	}

	if len(updates) == 0 {
		return Result{Message: "⚠️ *NO UPDATES FOUND*\n_Please specify what you want to change (e.g., 'name: John' or 'departure: tomorrow')._"}
	}

	// 3. Apply Updates
	var updatedFields []string
	departureUpdated := false
	var newDeparture time.Time
	arrivalExplicitlyUpdated := false
	var newArrival time.Time

	for field, value := range updates {
		// Strict Policy: Weight is fixed
		if field == "weight" {
			continue
		}

		// Validation
		if strings.Contains(field, "email") && !parser.ValidateEmail(value) {
			continue
		}
		if strings.Contains(field, "phone") && !parser.ValidatePhone(value) {
			continue
		}

		// Special Date Parsing — resolve timezone from shipment's origin country
		if field == "scheduled_transit_time" || field == "expected_delivery_time" || field == "outfordelivery_time" {
			originShip, _ := shipUC.Track(ctx, companyID, trackingID)
			var loc *time.Location
			if originShip != nil && originShip.Origin.Valid {
				originTZ := shipUC.GetService().ResolveTimezone(originShip.Origin.String)
				loc, _ = time.LoadLocation(originTZ)
			}
			if loc == nil {
				loc = time.UTC
			}
			now := time.Now().In(loc)

			parsedDate, ok := utils.ParseNaturalDate(value, now, loc)
			if !ok {
				logger.Warn().Str("field", field).Str("value", value).Msg("Failed to parse date in edit command")
				continue
			}

			value = parsedDate.UTC().Format("2006-01-02 15:04:05")
			if field == "scheduled_transit_time" {
				departureUpdated = true
				newDeparture = parsedDate.UTC()
			}

			if field == "expected_delivery_time" {
				arrivalExplicitlyUpdated = true
				newArrival = parsedDate.UTC()
			}
		}

		err := shipUC.UpdateField(ctx, companyID, trackingID, field, value)
		if err != nil {
			logger.Error().Err(err).Str("field", field).Msg("Failed to update field in edit command")
		} else {
			updatedFields = append(updatedFields, strings.ToUpper(strings.ReplaceAll(field, "_", " ")))
		}
	}

	if len(updatedFields) == 0 {
		return Result{Message: "⚠️ *UPDATE FAILED*\n_None of the fields could be updated. Check your format (e.g., label: value)._"}
	}

	// 4. Automatic Arrival/OFD Sync (Strict +1 Logic)
	// If departure was edited, but arrival was NOT explicitly edited in this command,
	// we auto-sync arrival to be exactly Departure + 1 day.
	if departureUpdated && !arrivalExplicitlyUpdated {
		dbShip, _ := shipUC.Track(ctx, companyID, trackingID)
		if dbShip != nil {
			destTZ := shipUC.GetService().ResolveTimezone(dbShip.Destination.String)
			destLoc, _ := time.LoadLocation(destTZ)
			if destLoc == nil {
				destLoc = time.UTC
			}

			// Add exactly 1 day
			arrivalDate := newDeparture.In(destLoc).AddDate(0, 0, 1)
			ofd := time.Date(arrivalDate.Year(), arrivalDate.Month(), arrivalDate.Day(), 8, 30, 0, 0, destLoc).UTC()
			arrival := time.Date(arrivalDate.Year(), arrivalDate.Month(), arrivalDate.Day(), 14, 0, 0, 0, destLoc).UTC()

			_ = shipUC.UpdateField(ctx, companyID, trackingID, "expected_delivery_time", arrival.Format("2006-01-02 15:04:05"))
			_ = shipUC.UpdateField(ctx, companyID, trackingID, "outfordelivery_time", ofd.Format("2006-01-02 15:04:05"))
			updatedFields = append(updatedFields, "ARRIVAL (AUTO-SYNC +1D)")
		}
	} else if arrivalExplicitlyUpdated {
		// Arrival edited -> Auto-sync OutForDelivery to 4 hours before arrival for consistency
		_ = shipUC.UpdateField(ctx, companyID, trackingID, "outfordelivery_time", newArrival.Add(-4*time.Hour).Format("2006-01-02 15:04:05"))
		updatedFields = append(updatedFields, "OFD (CONSISTENCY-SYNC)")
	}

	// 5. Persistence & Schedule Sync (Resolve Status based on new dates)
	// We ensure the shipment's STATUS matches the updated timeline.
	dbShip, _ := shipUC.Track(ctx, companyID, trackingID)
	if dbShip != nil {
		// Resolve status
		s := shipment.Shipment{
			Status: dbShip.Status.String,
		}
		if dbShip.ScheduledTransitTime.Valid {
			s.ScheduledTransitTime = &dbShip.ScheduledTransitTime.Time
		}
		if dbShip.OutfordeliveryTime.Valid {
			s.OutForDeliveryTime = &dbShip.OutfordeliveryTime.Time
		}
		if dbShip.ExpectedDeliveryTime.Valid {
			s.ExpectedDeliveryTime = &dbShip.ExpectedDeliveryTime.Time
		}

		newStatus := s.ResolveStatus(time.Now().UTC())
		if newStatus != dbShip.Status.String {
			_ = shipUC.UpdateField(ctx, companyID, trackingID, "status", newStatus)
			updatedFields = append(updatedFields, "STATUS (AUTO-SYNC)")

			// Trigger status alert if it changed
			if h.Sender != nil && h.Sender.GetWAClient() != nil {
				notif.SendStatusAlert(ctx, h.Sender, h.Cfg, h.CompanyName, dbShip.UserJid, trackingID, newStatus, dbShip.RecipientEmail.String, dbShip.Destination.String)
			}
		}
	}

	summary := fmt.Sprintf("✅ *INFORMATION UPDATED*\n\n🆔 *%s*\n\n📝 *FIELDS MODIFIED:*\n• %s\n\n━━━━━━━━━━━━━━━━━━━━━━━\n_Updates have been successfully persisted to the cloud._",
		trackingID, strings.Join(updatedFields, "\n• "))

	return Result{
		Message: summary,
		EditID:  trackingID,
	}
}
