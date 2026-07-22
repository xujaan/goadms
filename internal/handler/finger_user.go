package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jan/goadms/internal/model"
	"github.com/jan/goadms/internal/service"
)

type FingerUserHandler struct {
	fingerSvc *service.FingerUserService
}

func NewFingerUserHandler(fingerSvc *service.FingerUserService) *FingerUserHandler {
	return &FingerUserHandler{fingerSvc: fingerSvc}
}

// List handles GET /api/v1/fingerprint-users.
func (h *FingerUserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.fingerSvc.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if users == nil {
		users = []model.FingerprintUser{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": users})
}

// Create handles POST /api/v1/fingerprint-users.
func (h *FingerUserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var u model.FingerprintUser
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if u.EmployeeCode == "" || u.FullName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "employee_code and full_name required"})
		return
	}
	u.IsActive = true
	if err := h.fingerSvc.Create(r.Context(), &u); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

// Update handles PUT /api/v1/fingerprint-users/{id}.
func (h *FingerUserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := urlParamUUID(r, "id")
	var u model.FingerprintUser
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	u.ID = id
	if err := h.fingerSvc.Update(r.Context(), &u); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// Delete handles DELETE /api/v1/fingerprint-users/{id}.
func (h *FingerUserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := urlParamUUID(r, "id")
	if err := h.fingerSvc.Delete(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// SyncToDevice handles POST /api/v1/fingerprint-users/{id}/sync.
func (h *FingerUserHandler) SyncToDevice(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse(r.PathValue("id"))
	var req struct {
		DeviceID string `json:"device_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	deviceID, _ := uuid.Parse(req.DeviceID)

	if err := h.fingerSvc.SyncToDevice(r.Context(), userID, deviceID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "synced"})
}

// SyncAllToDevice handles POST /api/v1/fingerprint-users/sync-all.
func (h *FingerUserHandler) SyncAllToDevice(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DeviceID string `json:"device_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	deviceID, _ := uuid.Parse(req.DeviceID)

	count, err := h.fingerSvc.SyncAllToDevice(r.Context(), deviceID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"synced": count})
}
