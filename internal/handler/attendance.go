package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jan/goadms/internal/repository"
)

type AttendanceHandler struct {
	attendanceRepo *repository.AttendanceRepo
}

func NewAttendanceHandler(attendanceRepo *repository.AttendanceRepo) *AttendanceHandler {
	return &AttendanceHandler{attendanceRepo: attendanceRepo}
}

// List handles GET /api/v1/attendances.
func (h *AttendanceHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := repository.AttendanceFilter{
		Limit:  50,
		Offset: 0,
	}

	if sn := r.URL.Query().Get("device_sn"); sn != "" {
		filter.DeviceSN = sn
	}
	if emp := r.URL.Query().Get("employee_id"); emp != "" {
		filter.EmployeeID = emp
	}
	if src := r.URL.Query().Get("source"); src != "" {
		filter.Source = src
	}
	if df := r.URL.Query().Get("date_from"); df != "" {
		t, err := time.Parse("2006-01-02", df)
		if err == nil {
			filter.DateFrom = t
		}
	}
	if dt := r.URL.Query().Get("date_to"); dt != "" {
		t, err := time.Parse("2006-01-02", dt)
		if err == nil {
			filter.DateTo = t.Add(24*time.Hour).Add(-time.Second)
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			filter.Limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil {
			filter.Offset = n
		}
	}

	records, total, err := h.attendanceRepo.List(r.Context(), filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":   records,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// Delete handles DELETE /api/v1/attendances/{id}.
func (h *AttendanceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	// chi URLParam for this
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid attendance id"})
		return
	}
	if err := h.attendanceRepo.Delete(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
