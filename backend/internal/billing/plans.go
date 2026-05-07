package billing

import (
	"fmt"

	"webtracker-bot/internal/config"

	"github.com/google/uuid"
)

// Plan is the single source of truth for all pricing and quota data.
// The frontend fetches this via GET /api/billing/plans.
type Plan struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	NameKey      string   `json:"name_key"`
	DescKey      string   `json:"desc_key"`
	Price        int      `json:"price"`        // in kobo (subunits)
	Currency     string   `json:"currency"`     // e.g. "NGN"
	IntervalKey  string   `json:"interval_key"`
	Popular      bool     `json:"popular"`
	TrialKey     string   `json:"trial_key,omitempty"`
	BtnKey       string   `json:"btn_key"`
	Features     []string `json:"features"`      // Translation keys
	MaxShipments int64    `json:"max_shipments"`
}

var (
	PlanTrial = Plan{
		ID:           "trial",
		Name:         "Trial",
		MaxShipments: 50,
	}
	PlanStarter = Plan{
		ID:           "starter",
		Name:         "Starter",
		NameKey:      "starterName",
		DescKey:      "starterDesc",
		Price:        1200000, // ₦12,000
		Currency:     "NGN",
		IntervalKey:  "monthlyInterval",
		Popular:      false,
		TrialKey:     "sevenDayTrial",
		BtnKey:       "btnStartTrial",
		Features:     []string{"feat_50_shipments", "feat_whatsapp", "feat_web_portal", "feat_manual_entry", "feat_community"},
		MaxShipments: 50,
	}
	PlanPro = Plan{
		ID:           "pro",
		Name:         "Pro",
		NameKey:      "proName",
		DescKey:      "proDesc",
		Price:        3000000, // ₦30,000
		Currency:     "NGN",
		IntervalKey:  "monthlyInterval",
		Popular:      true,
		BtnKey:       "btnUpgradePro",
		Features:     []string{"feat_250_shipments", "feat_whatsapp", "feat_ai_parser", "feat_csv_upload", "feat_custom_branding", "feat_priority_support"},
		MaxShipments: 250,
	}
	PlanScale = Plan{
		ID:           "enterprise",
		Name:         "Scale",
		NameKey:      "scaleName",
		DescKey:      "scaleDesc",
		Price:        8500000, // ₦85,000
		Currency:     "NGN",
		IntervalKey:  "monthlyInterval",
		Popular:      false,
		BtnKey:       "btnContactSales",
		Features:     []string{"feat_1000_shipments", "feat_all_pro", "feat_api_webhook", "feat_dedicated_whatsapp", "feat_247_support"},
		MaxShipments: 1000,
	}
)

func GetPlans() []Plan {
	return []Plan{PlanStarter, PlanPro, PlanScale}
}

func GetPlanByID(id string) (Plan, error) {
	switch id {
	case "trial":
		return PlanTrial, nil
	case "starter":
		return PlanStarter, nil
	case "pro":
		return PlanPro, nil
	case "enterprise", "scale":
		return PlanScale, nil
	default:
		return Plan{}, fmt.Errorf("unknown plan: %s", id)
	}
}

// IsSuperAdmin checks if the given company ID matches the configured super admin.
// Lives in the billing package because it governs billing bypass logic.
func IsSuperAdmin(cfg *config.Config, companyID uuid.UUID) bool {
	if cfg == nil || cfg.SuperAdminCompanyID == "" {
		return false
	}
	superID, err := uuid.Parse(cfg.SuperAdminCompanyID)
	if err != nil {
		return false
	}
	return companyID == superID
}
