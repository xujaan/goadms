package webhook

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jan/goadms/internal/model"
	"github.com/jan/goadms/internal/repository"
)

// Deliverer sends HTTP POST to webhook URLs with retry logic.
type Deliverer struct {
	webhookRepo *repository.WebhookRepo
	signer      *Signer
	client      *http.Client
	retryMax    int
	retryBase   time.Duration
	logger      *slog.Logger
}

func NewDeliverer(webhookRepo *repository.WebhookRepo, signer *Signer, timeout time.Duration, retryMax int, retryBase time.Duration, logger *slog.Logger) *Deliverer {
	return &Deliverer{
		webhookRepo: webhookRepo,
		signer:      signer,
		client:      &http.Client{Timeout: timeout},
		retryMax:    retryMax,
		retryBase:   retryBase,
		logger:      logger,
	}
}

// Send delivers a webhook payload asynchronously. It logs the attempt to DB and retries on failure.
func (d *Deliverer) Send(ctx context.Context, cfg model.WebhookConfig, eventType string, payload []byte) {
	delivery := &model.WebhookDelivery{
		WebhookConfigID: cfg.ID,
		EventType:       eventType,
		Payload:         payload,
		RequestURL:      cfg.URL,
		Status:          "pending",
		AttemptCount:    0,
		MaxAttempts:     d.retryMax,
	}

	if err := d.webhookRepo.InsertDelivery(ctx, delivery); err != nil {
		d.logger.Error("webhook: insert delivery", "error", err)
		return
	}

	// Attempt delivery with retries.
	for attempt := 1; attempt <= d.retryMax; attempt++ {
		status, respBody, headers, duration, err := d.tryRequest(ctx, cfg, payload)

		delivery.AttemptCount = attempt
		delivery.DurationMs = int(duration.Milliseconds())
		delivery.RequestHeaders = headers

		if err == nil && status >= 200 && status < 300 {
			delivery.Status = "success"
			delivery.ResponseStatus = status
			delivery.ResponseBody = respBody
			d.webhookRepo.UpdateDelivery(ctx, delivery)
			return
		}

		delivery.ResponseStatus = status
		delivery.ResponseBody = respBody
		if err != nil {
			delivery.ErrorMessage = err.Error()
		} else {
			delivery.ErrorMessage = fmt.Sprintf("non-2xx response: %d", status)
		}

		if attempt < d.retryMax {
			delivery.Status = "retrying"
			next := time.Now().Add(d.retryBase * time.Duration(1<<(attempt-1))) // exponential backoff
			delivery.NextAttemptAt = &next
			d.webhookRepo.UpdateDelivery(ctx, delivery)
			d.logger.Warn("webhook: retry", "url", cfg.URL, "attempt", attempt, "error", delivery.ErrorMessage)
			time.Sleep(d.retryBase * time.Duration(1<<(attempt-1)))
		} else {
			delivery.Status = "failed"
			delivery.NextAttemptAt = nil
			d.webhookRepo.UpdateDelivery(ctx, delivery)
			d.logger.Error("webhook: failed", "url", cfg.URL, "attempts", attempt, "error", delivery.ErrorMessage)
		}
	}
}

func (d *Deliverer) tryRequest(ctx context.Context, cfg model.WebhookConfig, payload []byte) (status int, body string, headers map[string]string, dur time.Duration, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.URL, bytes.NewReader(payload))
	if err != nil {
		return 0, "", nil, 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "goadms-webhook/1.0")
	req.Header.Set("X-ADMS-Event", "")

	// Custom headers from config.
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	// HMAC signature.
	if cfg.Secret != "" {
		sig := d.signer.Sign(payload, cfg.Secret)
		req.Header.Set("X-ADMS-Signature", sig)
	}

	headers = make(map[string]string)
	for k := range req.Header {
		headers[k] = req.Header.Get(k)
	}

	start := time.Now()
	resp, err := d.client.Do(req)
	dur = time.Since(start)
	if err != nil {
		return 0, "", headers, dur, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	return resp.StatusCode, buf.String(), headers, dur, nil
}
