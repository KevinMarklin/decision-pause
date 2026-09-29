package handler

import (
	"net/http"

	"github.com/KevinMarklin/decision-pause/backend/internal/calculator"
	"github.com/KevinMarklin/decision-pause/backend/internal/model"
)

func (h *Handler) analyzeMarketing(w http.ResponseWriter, r *http.Request) {
	var in model.MarketingInput
	if e := decodeBody(w, r, &in); e != nil {
		writeProtoError(w, e)
		return
	}
	if fields := validateMarketing(in); len(fields) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation", "fields": fields})
		return
	}
	writeJSON(w, http.StatusOK, calculator.MarketingResult(in))
}

func validateMarketing(in model.MarketingInput) map[string]string {
	fields := map[string]string{}
	if in.Budget <= 0 { fields["budget"] = "должен быть больше 0" }
	if in.AvgCheck <= 0 { fields["avg_check"] = "должен быть больше 0" }
	if in.MarginPct <= 0 || in.MarginPct > 100 { fields["margin_pct"] = "от 0 до 100%" }
	if in.LeadCost <= 0 { fields["lead_cost"] = "должна быть больше 0" }
	if in.ExpectedBoostPct < -100 || in.ExpectedBoostPct > 1000 { fields["expected_boost_pct"] = "от -100 до 1000%" }
	if in.CurrentRevenue < 0 { fields["current_revenue"] = "не может быть отрицательной" }
	if in.FixedExpenses < 0 { fields["fixed_expenses"] = "не могут быть отрицательными" }
	if in.Reserve < 0 { fields["reserve"] = "не может быть отрицательным" }
	if len(fields) == 0 { return nil }
	return fields
}
