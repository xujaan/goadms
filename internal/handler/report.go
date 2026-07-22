package handler

import (
	"net/http"
	"time"

	"github.com/jan/goadms/internal/service"
)

type ReportHandler struct {
	reportSvc *service.ReportService
}

func NewReportHandler(reportSvc *service.ReportService) *ReportHandler {
	return &ReportHandler{reportSvc: reportSvc}
}

// AttendanceReport handles GET /api/v1/reports/attendance.
func (h *ReportHandler) AttendanceReport(w http.ResponseWriter, r *http.Request) {
	filter := parseReportFilter(r)

	result, err := h.reportSvc.GenerateAttendanceReport(r.Context(), filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// AttendanceCSV handles GET /api/v1/reports/attendance.csv.
func (h *ReportHandler) AttendanceCSV(w http.ResponseWriter, r *http.Request) {
	filter := parseReportFilter(r)

	csv, err := h.reportSvc.GenerateCSV(r.Context(), filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=attendance_report.csv")
	w.Write([]byte(csv))
}

func parseReportFilter(r *http.Request) service.ReportFilter {
	now := time.Now()
	f := service.ReportFilter{
		DateFrom: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		DateTo:   now,
	}

	if df := r.URL.Query().Get("date_from"); df != "" {
		t, err := time.Parse("2006-01-02", df)
		if err == nil {
			f.DateFrom = t
		}
	}
	if dt := r.URL.Query().Get("date_to"); dt != "" {
		t, err := time.Parse("2006-01-02", dt)
		if err == nil {
			f.DateTo = t
		}
	}
	if sn := r.URL.Query().Get("device_sn"); sn != "" {
		f.DeviceSN = sn
	}
	return f
}
