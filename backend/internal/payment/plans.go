package payment

import "fmt"

// Plan is the single source of truth for all pricing data.
// The frontend fetches this via GET /api/billing/plans.
type Plan struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	NameKey     string   `json:"name_key"`
	DescKey     string   `json:"desc_key"`
	Price       int      `json:"price"`       // in kobo (subunits)
	Currency    string   `json:"currency"`    // e.g. "NGN"
	IntervalKey string   `json:"interval_key"`
	Popular     bool     `json:"popular"`
	TrialKey    string   `json:"trial_key,omitempty"`
	BtnKey      string   `json:"btn_key"`
	Features    []string `json:"features"` // Translation keys
}

var (
	PlanStarter = Plan{
		ID:          "starter",
		Name:        "Starter",
		NameKey:     "starterName",
		DescKey:     "starterDesc",
		Price:       1200000, // ₦12,000
		Currency:    "NGN",
		IntervalKey: "monthlyInterval",
		Popular:     false,
		TrialKey:    "sevenDayTrial",
		BtnKey:      "btnStartTrial",
		Features:    []string{"feat_50_shipments", "feat_whatsapp", "feat_web_portal", "feat_manual_entry", "feat_community"},
	}
	PlanPro = Plan{
		ID:          "pro",
		Name:        "Pro",
		NameKey:     "proName",
		DescKey:     "proDesc",
		Price:       3000000, // ₦30,000
		Currency:    "NGN",
		IntervalKey: "monthlyInterval",
		Popular:     true,
		BtnKey:      "btnUpgradePro",
		Features:    []string{"feat_250_shipments", "feat_whatsapp", "feat_ai_parser", "feat_csv_upload", "feat_custom_branding", "feat_priority_support"},
	}
	PlanScale = Plan{
		ID:          "enterprise",
		Name:        "Scale",
		NameKey:     "scaleName",
		DescKey:     "scaleDesc",
		Price:       8500000, // ₦85,000
		Currency:    "NGN",
		IntervalKey: "monthlyInterval",
		Popular:     false,
		BtnKey:      "btnContactSales",
		Features:    []string{"feat_1000_shipments", "feat_all_pro", "feat_api_webhook", "feat_dedicated_whatsapp", "feat_247_support"},
	}
)

func GetPlans() []Plan {
	return []Plan{PlanStarter, PlanPro, PlanScale}
}

func GetPlanByID(id string) (Plan, error) {
	switch id {
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
