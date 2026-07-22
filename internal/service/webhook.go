package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jan/goadms/internal/model"
	"github.com/jan/goadms/internal/repository"
)

type WebhookService struct {
	webhookRepo *repository.WebhookRepo
}

func NewWebhookService(webhookRepo *repository.WebhookRepo) *WebhookService {
	return &WebhookService{webhookRepo: webhookRepo}
}

func (s *WebhookService) ListByDevice(ctx context.Context, deviceID uuid.UUID) ([]model.WebhookConfig, error) {
	return s.webhookRepo.ListByDevice(ctx, deviceID)
}

func (s *WebhookService) Create(ctx context.Context, c *model.WebhookConfig) error {
	return s.webhookRepo.CreateConfig(ctx, c)
}

func (s *WebhookService) Update(ctx context.Context, c *model.WebhookConfig) error {
	return s.webhookRepo.UpdateConfig(ctx, c)
}

func (s *WebhookService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.webhookRepo.DeleteConfig(ctx, id)
}

func (s *WebhookService) GetConfig(ctx context.Context, id uuid.UUID) (*model.WebhookConfig, error) {
	return s.webhookRepo.GetConfigByID(ctx, id)
}

func (s *WebhookService) Deliveries(ctx context.Context, configID uuid.UUID, status string, limit int) ([]model.WebhookDelivery, error) {
	return s.webhookRepo.ListDeliveriesByConfig(ctx, configID, status, limit)
}

func (s *WebhookService) GetDelivery(ctx context.Context, id int64) (*model.WebhookDelivery, error) {
	return s.webhookRepo.GetDeliveryByID(ctx, id)
}
