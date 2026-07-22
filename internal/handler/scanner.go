package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jan/goadms/internal/service"
)

type ScannerHandler struct {
	scannerSvc *service.ScannerService
}

func NewScannerHandler(scannerSvc *service.ScannerService) *ScannerHandler {
	return &ScannerHandler{scannerSvc: scannerSvc}
}

// ScanSubnet handles POST /api/v1/detect/scan.
func (h *ScannerHandler) ScanSubnet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Subnet string `json:"subnet"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Subnet = r.URL.Query().Get("subnet")
	}
	if req.Subnet == "" {
		req.Subnet = "192.168.1.0/24"
	}

	results, err := h.scannerSvc.ScanSubnet(r.Context(), req.Subnet)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Filter only open ports.
	var open []service.ScanResult
	for _, r := range results {
		if r.Open {
			open = append(open, r)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"subnet":  req.Subnet,
		"total":   len(results),
		"open":    len(open),
		"results": open,
	})
}

// DetectSingle handles POST /api/v1/detect/one.
func (h *ScannerHandler) DetectSingle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IP string `json:"ip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.IP = r.URL.Query().Get("ip")
	}
	if req.IP == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ip is required"})
		return
	}

	result := h.scannerSvc.DetectSingle(req.IP)
	writeJSON(w, http.StatusOK, result)
}
