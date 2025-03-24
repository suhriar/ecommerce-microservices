package rest

import (
	"encoding/json"
	"net/http"
	"pricing-service/internal/usecase"
	"pricing-service/pkg/utils"
)

type PricingHandler struct {
	pricingUsecase usecase.PricingUsecase
}

func NewPricingHandler(pricingUsecase usecase.PricingUsecase) *PricingHandler {
	return &PricingHandler{pricingUsecase: pricingUsecase}
}

func (h *PricingHandler) GetPricing(w http.ResponseWriter, r *http.Request) {
	var pricingRequest struct {
		ProductID int `json:"product_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&pricingRequest); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	pricing, err := h.pricingUsecase.CalculatePricing(r.Context(), pricingRequest.ProductID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, pricing)
}
