package webhook

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jan/goadms/internal/model"
	"github.com/jan/goadms/internal/repository"
)

// Dispatcher fans out events to matching webhook configs.
type Dispatcher struct {
	webhookRepo *repository.WebhookRepo
	deliverer   *Deliverer
	logger      *slog.Logger
}

func NewDispatcher(webhookRepo *repository.WebhookRepo, deliverer *Deliverer, logger *slog.Logger) *Dispatcher {
	return &Dispatcher{
		webhookRepo: webhookRepo,
		deliverer:   deliverer,
		logger:      logger,
	}
}

// FanOut finds all active webhooks matching the event for the device and dispatches them.
func (d *Dispatcher) FanOut(ctx context.Context, eventType string, deviceID uuid.UUID, deviceSN string, data any) {
	go d.dispatchAsync(context.Background(), eventType, deviceID, deviceSN, data)
}

// FanOutAll finds all active webhooks globally (not scoped to device) matching the event.
func (d *Dispatcher) FanOutAll(ctx context.Context, eventType string, deviceID uuid.UUID, deviceSN string, data any) {
	go d.dispatchAllAsync(context.Background(), eventType, deviceID, deviceSN, data)
}

func (d *Dispatcher) dispatchAsync(ctx context.Context, eventType string, deviceID uuid.UUID, deviceSN string, data any) {
	configs, err := d.webhookRepo.ListActiveByDeviceAndEvent(ctx, deviceID, eventType)
	if err != nil {
		d.logger.Error("webhook dispatch: list configs", "error", err, "device_id", deviceID)
		return
	}

	d.dispatchConfigs(ctx, eventType, deviceID, deviceSN, data, configs)
}

func (d *Dispatcher) dispatchAllAsync(ctx context.Context, eventType string, deviceID uuid.UUID, deviceSN string, data any) {
	configs, err := d.webhookRepo.GetAllActiveByEvent(ctx, eventType)
	if err != nil {
		d.logger.Error("webhook dispatch all: list configs", "error", err)
		return
	}

	d.dispatchConfigs(ctx, eventType, deviceID, deviceSN, data, configs)
}

func (d *Dispatcher) dispatchConfigs(ctx context.Context, eventType string, deviceID uuid.UUID, deviceSN string, data any, configs []model.WebhookConfig) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		d.logger.Error("webhook dispatch: marshal data", "error", err)
		return
	}

	event := model.WebhookEvent{
		Event:     eventType,
		DeviceID:  deviceID,
		DeviceSN:  deviceSN,
		Timestamp: time.Now(),
		Data:      dataBytes,
	}
	payload, _ := json.Marshal(event)

	for _, cfg := range configs {
		d.deliverer.Send(ctx, cfg, eventType, payload)
	}
}
