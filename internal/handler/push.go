package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jan/goadms/internal/model"
	"github.com/jan/goadms/internal/repository"
	"github.com/jan/goadms/internal/service"
)

type PushHandler struct {
	deviceSvc *service.DeviceService
	pushSvc   *service.PushService
	rawLogRepo *repository.RawLogRepo
}

func NewPushHandler(deviceSvc *service.DeviceService, pushSvc *service.PushService, rawLogRepo *repository.RawLogRepo) *PushHandler {
	return &PushHandler{
		deviceSvc:  deviceSvc,
		pushSvc:    pushSvc,
		rawLogRepo: rawLogRepo,
	}
}

// Handshake handles GET /iclock/cdata.
func (h *PushHandler) Handshake(w http.ResponseWriter, r *http.Request) {
	sn := r.URL.Query().Get("SN")
	ip := clientIP(r)

	// Record handshake — auto-registers device if new.
	if err := h.deviceSvc.RecordHandshake(r.Context(), sn, ip); err != nil {
		// Don't fail — device should still get response.
	}

	// Build handshake config (from device or defaults).
	cfg := model.DefaultHandshakeConfig()
	device, _ := h.deviceSvc.GetBySN(r.Context(), sn)
	if device != nil && device.HandshakeConfig.Stamp > 0 {
		cfg = device.HandshakeConfig
	}

	response := fmt.Sprintf(
		"OPTIONS:\r\n"+
			"Stamp=%d\r\n"+
			"OpStamp=%d\r\n"+
			"ErrorDelay=%d\r\n"+
			"Delay=%d\r\n"+
			"TransTimes=%s\r\n"+
			"TransInterval=%d\r\n"+
			"TransFlag=%s\r\n"+
			"TimeZone=%d\r\n"+
			"Realtime=%d\r\n"+
			"Encrypt=%d\r\n",
		cfg.Stamp, cfg.OpStamp, cfg.ErrorDelay, cfg.Delay,
		cfg.TransTimes, cfg.TransInterval, cfg.TransFlag,
		cfg.TimeZone, cfg.Realtime, cfg.Encrypt,
	)

	// Log raw request.
	h.rawLogRepo.Insert(r.Context(), &model.RawLog{
		DeviceSN:      sn,
		RequestMethod: "GET",
		RequestURI:    r.URL.RequestURI(),
		ResponseBody:  response,
		LogType:       "push_handshake",
	})

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(response))
}

// ReceiveRecords handles POST /iclock/cdata.
func (h *PushHandler) ReceiveRecords(w http.ResponseWriter, r *http.Request) {
	sn := r.URL.Query().Get("SN")
	table := r.URL.Query().Get("table")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "ERROR: cannot read body", http.StatusBadRequest)
		return
	}
	bodyStr := string(body)

	// Process and store records.
	count, err := h.pushSvc.ProcessRecords(r.Context(), sn, bodyStr, table)
	if err != nil {
		// Log the error.
		h.rawLogRepo.Insert(r.Context(), &model.RawLog{
			DeviceSN:      sn,
			RequestMethod: "POST",
			RequestURI:    r.URL.RequestURI(),
			RequestBody:   bodyStr,
			ResponseBody:  fmt.Sprintf("ERROR: %v", err),
			LogType:       "push_records",
		})
		http.Error(w, fmt.Sprintf("ERROR: %d", count), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("OK: %d", count)))
}

// Test handles GET /iclock/test — debug endpoint that logs raw request.
func (h *PushHandler) Test(w http.ResponseWriter, r *http.Request) {
	sn := r.URL.Query().Get("SN")
	body, _ := io.ReadAll(r.Body)

	h.rawLogRepo.Insert(r.Context(), &model.RawLog{
		DeviceSN:      sn,
		RequestMethod: "GET",
		RequestURI:    r.URL.RequestURI(),
		RequestBody:   string(body),
		LogType:       "push_test",
	})

	w.Write([]byte("OK"))
}

// GetRequest handles GET /iclock/getrequest — simple ack.
func (h *PushHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	sn := r.URL.Query().Get("SN")
	h.rawLogRepo.Insert(r.Context(), &model.RawLog{
		DeviceSN:      sn,
		RequestMethod: "GET",
		RequestURI:    r.URL.RequestURI(),
		LogType:       "push_getrequest",
	})

	w.Write([]byte("OK"))
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Strip port from RemoteAddr.
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}
