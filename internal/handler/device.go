package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jan/goadms/internal/model"
	"github.com/jan/goadms/internal/service"
	"github.com/jan/goadms/internal/zkteco"
)

type DeviceHandler struct {
	deviceSvc  *service.DeviceService
	zktecoSvc  *service.ZkTecoService
}

func NewDeviceHandler(deviceSvc *service.DeviceService, zktecoSvc *service.ZkTecoService) *DeviceHandler {
	return &DeviceHandler{deviceSvc: deviceSvc, zktecoSvc: zktecoSvc}
}

// List handles GET /api/v1/devices.
func (h *DeviceHandler) List(w http.ResponseWriter, r *http.Request) {
	devices, err := h.deviceSvc.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if devices == nil {
		devices = []model.Device{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": devices})
}

// Create handles POST /api/v1/devices.
func (h *DeviceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var d model.Device
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if d.Name == "" || d.SerialNumber == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and serial_number are required"})
		return
	}
	if d.Protocol == "" {
		d.Protocol = "zk-tcp"
	}
	if d.Port == 0 {
		d.Port = 4370
	}
	d.IsActive = true
	if d.Timezone == "" {
		d.Timezone = "Asia/Jakarta"
	}
	if d.HandshakeConfig.Stamp == 0 {
		d.HandshakeConfig = model.DefaultHandshakeConfig()
	}
	if err := h.deviceSvc.Create(r.Context(), &d); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, d)
}

// GetByID handles GET /api/v1/devices/{id}.
func (h *DeviceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamUUID(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid device id"})
		return
	}
	d, err := h.deviceSvc.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if d == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "device not found"})
		return
	}
	writeJSON(w, http.StatusOK, d)
}

// Update handles PUT /api/v1/devices/{id}.
func (h *DeviceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamUUID(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid device id"})
		return
	}
	var d model.Device
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	d.ID = id
	if err := h.deviceSvc.Update(r.Context(), &d); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, d)
}

// Delete handles DELETE /api/v1/devices/{id}.
func (h *DeviceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamUUID(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid device id"})
		return
	}
	if err := h.deviceSvc.Delete(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "device deleted"})
}

// --- TCP operations ---

// TestConnection handles POST /api/v1/devices/{id}/test-connection.
func (h *DeviceHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamUUID(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid device id"})
		return
	}
	if err := h.zktecoSvc.TestConnection(r.Context(), id); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "message": "connection successful"})
}

// PullAttendance handles POST /api/v1/devices/{id}/pull-attendance.
func (h *DeviceHandler) PullAttendance(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamUUID(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid device id"})
		return
	}
	count, err := h.zktecoSvc.PullAttendance(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"records_pulled": count})
}

// ListDeviceUsers handles GET /api/v1/devices/{id}/users.
func (h *DeviceHandler) ListDeviceUsers(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamUUID(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid device id"})
		return
	}
	users, err := h.zktecoSvc.GetDeviceUsers(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if users == nil {
		users = []zkteco.UserRecord{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": users})
}

// Reboot handles POST /api/v1/devices/{id}/reboot.
func (h *DeviceHandler) Reboot(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamUUID(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid device id"})
		return
	}
	if err := h.zktecoSvc.RebootDevice(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "reboot command sent"})
}

// SyncTime handles POST /api/v1/devices/{id}/sync-time.
func (h *DeviceHandler) SyncTime(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamUUID(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid device id"})
		return
	}
	if err := h.zktecoSvc.SyncDeviceTime(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "time synced"})
}

// DeleteDeviceUser handles POST /api/v1/devices/{id}/users/{uid}/delete.
func (h *DeviceHandler) DeleteDeviceUser(w http.ResponseWriter, r *http.Request) {
	id, _ := urlParamUUID(r, "id")
	uid := r.PathValue("uid")
	if err := h.zktecoSvc.DeleteDeviceUser(r.Context(), id, uid); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
