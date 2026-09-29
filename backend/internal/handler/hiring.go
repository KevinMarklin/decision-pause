package handler

import (
	"net/http"

	"github.com/KevinMarklin/decision-pause/backend/internal/calculator"
	"github.com/KevinMarklin/decision-pause/backend/internal/model"
)

func (h *Handler) analyzeHiring(w http.ResponseWriter, r *http.Request) {
	var in model.HiringInput
	if e := decodeBody(w, r, &in); e != nil {
		writeProtoError(w, e)
		return
	}
	if fields := validateHiring(in); len(fields) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation", "fields": fields})
		return
	}
	writeJSON(w, http.StatusOK, calculator.CalculateHiring(in))
}

func validateHiring(in model.HiringInput) map[string]string {
	fields := map[string]string{}
	if in.Salary <= 0 { fields["salary"] = "должен быть больше 0" }
	if in.TaxRate < 0 || in.TaxRate > 1 { fields["tax_rate"] = "от 0 до 1 или от 0 до 100% в десятичном формате" }
	if in.OverheadCosts < 0 { fields["overhead_costs"] = "не могут быть отрицательными" }
	if in.MarginPercent <= 0 || in.MarginPercent > 100 { fields["margin_percent"] = "от 1 до 100%" }
	if in.AverageCheck < 0 { fields["average_check"] = "не может быть отрицательным" }
	if len(fields) == 0 { return nil }
	return fields
}
