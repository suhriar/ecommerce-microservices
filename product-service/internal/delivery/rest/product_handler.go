package rest

import (
	"encoding/json"
	"net/http"
	"product-service/internal/usecase"
	"product-service/pkg/utils"
	"strconv"

	"github.com/gorilla/mux"
)

type ProductHandler struct {
	productUsecase usecase.ProductUsecase
}

func NewproductHandler(productUsecase usecase.ProductUsecase) *ProductHandler {
	return &ProductHandler{productUsecase: productUsecase}
}

// GetProductStock gets the stock of a product --> /products/:id/stock
func (h *ProductHandler) GetProductStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		return
	}

	stock, err := h.productUsecase.GetProductStock(r.Context(), id)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]int{"stock": stock})
}

// ReserveProductStock reserves stock for a product --> /products/reserve
func (h *ProductHandler) ReserveProductStock(w http.ResponseWriter, r *http.Request) {
	reservation := struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reservation); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}
	err := h.productUsecase.ReserveProductStock(r.Context(), reservation.ProductID, reservation.Quantity)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Stock reserved"})
}

// ReleaseProductStock releases stock for a product --> /products/release
func (h *ProductHandler) ReleaseProductStock(w http.ResponseWriter, r *http.Request) {
	release := struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&release); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}
	err := h.productUsecase.ReleaseProductStock(r.Context(), release.ProductID, release.Quantity)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Stock released"})
}

// PreWarmupCache pre-warms the cache with product data --> /products/warmup-cache
func (h *ProductHandler) PreWarmupCache(w http.ResponseWriter, r *http.Request) {
	//// call synchronously
	//err := h.productUsecase.PreWarmCache(r.Context())
	//if err != nil {
	//	utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	//	return
	//}

	// call asynchrously
	err := h.productUsecase.PreWarmCacheAsync(r.Context())
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Cache pre-warmed"})
}
