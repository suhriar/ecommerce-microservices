package rest

import (
	"encoding/json"
	"net/http"
	"order-service/domain"
	"order-service/internal/usecase"
	"order-service/pkg/utils"
	"strconv"

	"github.com/gorilla/mux"
)

type OrderHandler struct {
	orderUsecase usecase.OrderUsecase
}

func NewOrderHandler(orderUsecase usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{orderUsecase: orderUsecase}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	order := domain.Order{}
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}
	order.IdempotentKey = r.Header.Get("Idempotent-Key")

	createdOrder, err := h.orderUsecase.CreateOrder(r.Context(), order)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, createdOrder)
}

func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	order := domain.Order{}
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	updatedOrder, err := h.orderUsecase.UpdateOrder(r.Context(), order)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, updatedOrder)
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		return
	}

	order, err := h.orderUsecase.CancelOrder(r.Context(), id)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, order)
}
