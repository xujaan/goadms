package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jan/goadms/internal/model"
	"github.com/jan/goadms/internal/service"
)

func mergeWebhook(dst *model.WebhookConfig, src *model.WebhookConfig) {
	if src.Name != "" { dst.Name = src.Name }
	if src.URL != "" { dst.URL = src.URL }
	if src.Secret != "" { dst.Secret = src.Secret }
	if len(src.Events) > 0 { dst.Events = src.Events }
	if src.Headers != nil && len(src.Headers) > 0 { dst.Headers = src.Headers }
	dst.IsActive = src.IsActive
}

type WebhookHandler struct {
	webhookSvc *service.WebhookService
}

func NewWebhookHandler(webhookSvc *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{webhookSvc: webhookSvc}
}

// List handles GET /api/v1/devices/{id}/webhooks.
func (h *WebhookHandler) List(w http.ResponseWriter, r *http.Request) {
	deviceID, err := urlParamUUID(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid device id"})
		return
	}
	configs, err := h.webhookSvc.ListByDevice(r.Context(), deviceID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if configs == nil {
		configs = []model.WebhookConfig{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": configs})
}

// Create handles POST /api/v1/devices/{id}/webhooks.
func (h *WebhookHandler) Create(w http.ResponseWriter, r *http.Request) {
	deviceID, err := urlParamUUID(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid device id"})
		return
	}
	var cfg model.WebhookConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	cfg.DeviceID = deviceID
	if cfg.Name == "" || cfg.URL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and url are required"})
		return
	}
	if len(cfg.Events) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "at least one event is required"})
		return
	}
	if err := h.webhookSvc.Create(r.Context(), &cfg); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, cfg)
}

// Update handles PUT /api/v1/devices/{id}/webhooks/{wid}.
func (h *WebhookHandler) Update(w http.ResponseWriter, r *http.Request) {
	_, err := urlParamUUID(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid device id"})
		return
	}
	webhookID, err := urlParamUUID(r, "wid")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid webhook id"})
		return
	}
	// Fetch existing webhook, then merge request fields.
	existing, err := h.webhookSvc.GetConfig(r.Context(), webhookID)
	if err != nil || existing == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "webhook not found"})
		return
	}
	var req model.WebhookConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	// Merge: hanya overwrite field yang dikirim (non-zero).
	mergeWebhook(existing, &req)
	existing.UpdatedAt = time.Now()
	if err := h.webhookSvc.Update(r.Context(), existing); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, existing)
}

// Delete handles DELETE /api/v1/devices/{id}/webhooks/{wid}.
func (h *WebhookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	webhookID, err := urlParamUUID(r, "wid")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid webhook id"})
		return
	}
	if err := h.webhookSvc.Delete(r.Context(), webhookID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "webhook deleted"})
}

// ListDeliveries handles GET /api/v1/devices/{id}/webhooks/{wid}/deliveries.
func (h *WebhookHandler) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	webhookID, err := urlParamUUID(r, "wid")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid webhook id"})
		return
	}
	status := queryParamStr(r, "status", "")
	limit := queryParamInt(r, "limit", 50)

	deliveries, err := h.webhookSvc.Deliveries(r.Context(), webhookID, status, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if deliveries == nil {
		deliveries = []model.WebhookDelivery{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": deliveries})
}

// TestPing handles POST /api/v1/devices/{id}/webhooks/{wid}/test.
func (h *WebhookHandler) TestPing(w http.ResponseWriter, r *http.Request) {
	webhookID, err := urlParamUUID(r, "wid")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid webhook id"})
		return
	}
	cfg, err := h.webhookSvc.GetConfig(r.Context(), webhookID)
	if err != nil || cfg == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "webhook not found"})
		return
	}
	// Test ping will be implemented when webhook deliverer is wired in.
	writeJSON(w, http.StatusOK, map[string]any{
		"message":       "test ping queued",
		"webhook_id":    cfg.ID,
		"webhook_url":   cfg.URL,
		"events":        cfg.Events,
		"is_active":     cfg.IsActive,
	})
}
