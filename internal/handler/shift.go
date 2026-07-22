package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jan/goadms/internal/model"
	"github.com/jan/goadms/internal/service"
)

type ShiftHandler struct {
	shiftSvc *service.ShiftService
}

func NewShiftHandler(shiftSvc *service.ShiftService) *ShiftHandler {
	return &ShiftHandler{shiftSvc: shiftSvc}
}

// List handles GET /api/v1/shifts.
func (h *ShiftHandler) List(w http.ResponseWriter, r *http.Request) {
	shifts, err := h.shiftSvc.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if shifts == nil {
		shifts = []model.Shift{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": shifts})
}

// Create handles POST /api/v1/shifts.
func (h *ShiftHandler) Create(w http.ResponseWriter, r *http.Request) {
	var s model.Shift
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if s.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if s.StartTime == "" {
		s.StartTime = "08:00"
	}
	if s.EndTime == "" {
		s.EndTime = "17:00"
	}
	s.IsActive = true
	if err := h.shiftSvc.Create(r.Context(), &s); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

// Update handles PUT /api/v1/shifts/{id}.
func (h *ShiftHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := urlParamUUID(r, "id")
	var s model.Shift
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	s.ID = id
	if err := h.shiftSvc.Update(r.Context(), &s); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// Delete handles DELETE /api/v1/shifts/{id}.
func (h *ShiftHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := urlParamUUID(r, "id")
	if err := h.shiftSvc.Delete(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// AssignUser handles POST /api/v1/shifts/{id}/assign.
func (h *ShiftHandler) AssignUser(w http.ResponseWriter, r *http.Request) {
	shiftID, _ := urlParamUUID(r, "id")
	var req struct {
		UserID        uuid.UUID `json:"user_id"`
		EffectiveDate string    `json:"effective_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if err := h.shiftSvc.AssignUser(r.Context(), shiftID, req.UserID, req.EffectiveDate); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "assigned"})
}

// GetAssignedUsers handles GET /api/v1/shifts/{id}/users.
func (h *ShiftHandler) GetAssignedUsers(w http.ResponseWriter, r *http.Request) {
	id, _ := urlParamUUID(r, "id")
	users, err := h.shiftSvc.GetAssignedUsers(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if users == nil {
		users = []model.UserShift{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": users})
}
