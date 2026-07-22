package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jan/goadms/internal/model"
	"github.com/jan/goadms/internal/repository"
)

type ReportService struct {
	attendanceRepo *repository.AttendanceRepo
	shiftRepo      *repository.ShiftRepo
	deviceRepo     *repository.DeviceRepo
}

func NewReportService(attendanceRepo *repository.AttendanceRepo, shiftRepo *repository.ShiftRepo, deviceRepo *repository.DeviceRepo) *ReportService {
	return &ReportService{
		attendanceRepo: attendanceRepo,
		shiftRepo:      shiftRepo,
		deviceRepo:     deviceRepo,
	}
}

type ReportFilter struct {
	DateFrom time.Time
	DateTo   time.Time
	DeviceSN string
}

type DailyReport struct {
	Date       string `json:"date"`
	EmployeeID string `json:"employee_id"`
	DeviceSN   string `json:"device_sn"`
	CheckIn    string `json:"check_in"`
	CheckOut   string `json:"check_out"`
	LateMin    int    `json:"late_minutes"`
	EarlyMin   int    `json:"early_minutes"`
	Overtime   int    `json:"overtime_minutes"`
	Status     string `json:"status"`
}

type ReportResult struct {
	Reports []DailyReport `json:"reports"`
	Summary ReportSummary `json:"summary"`
}

type ReportSummary struct {
	TotalRecords   int    `json:"total_records"`
	Hadir          int    `json:"hadir"`
	Terlambat      int    `json:"terlambat"`
	TidakHadir     int    `json:"tidak_hadir"`
	DateFrom       string `json:"date_from"`
	DateTo         string `json:"date_to"`
}

// GenerateAttendanceReport creates an attendance report grouped by (date, employee, device).
func (s *ReportService) GenerateAttendanceReport(ctx context.Context, filter ReportFilter) (*ReportResult, error) {
	// Get all attendance records within date range.
	records, _, err := s.attendanceRepo.List(ctx, repository.AttendanceFilter{
		DateFrom: filter.DateFrom,
		DateTo:   filter.DateTo.Add(24 * time.Hour).Add(-time.Second),
		DeviceSN: filter.DeviceSN,
		Limit:    50000,
	})
	if err != nil {
		return nil, fmt.Errorf("list attendances: %w", err)
	}

	// Group by (date, employee_id, device_sn).
	type groupKey struct {
		date       string
		employeeID string
		deviceSN   string
	}
	groups := make(map[groupKey][]model.Attendance)
	for _, r := range records {
		date := r.Timestamp.Format("2006-01-02")
		key := groupKey{date: date, employeeID: r.EmployeeID, deviceSN: r.DeviceSN}
		groups[key] = append(groups[key], r)
	}

	// Get all shifts for reference.
	shifts, _ := s.shiftRepo.List(ctx)
	defaultShift := model.Shift{
		StartTime:            "08:00",
		EndTime:              "17:00",
		BreakMinutes:         60,
		LateToleranceMinutes: 5,
	}

	var reports []DailyReport
	for key, atts := range groups {
		// Sort by timestamp.
		sort.Slice(atts, func(i, j int) bool {
			return atts[i].Timestamp.Before(atts[j].Timestamp)
		})

		checkIn := atts[0].Timestamp
		checkOut := atts[len(atts)-1].Timestamp

		// Use first matching shift or default.
		shift := defaultShift
		for _, s := range shifts {
			if s.IsActive {
				shift = s
				break
			}
		}

		shiftStart := parseTime(shift.StartTime)
		shiftEnd := parseTime(shift.EndTime)

		lateMin := 0
		if checkIn.After(shiftStart.Add(time.Duration(shift.LateToleranceMinutes) * time.Minute)) {
			lateMin = int(checkIn.Sub(shiftStart).Minutes())
		}

		earlyMin := 0
		if checkOut.Before(shiftEnd) {
			earlyMin = int(shiftEnd.Sub(checkOut).Minutes())
		}

		overtime := 0
		if shift.OvertimeAfterMinutes > 0 {
			overtimeThreshold := shiftEnd.Add(time.Duration(shift.OvertimeAfterMinutes) * time.Minute)
			if checkOut.After(overtimeThreshold) {
				overtime = int(checkOut.Sub(shiftEnd).Minutes())
			}
		}

		status := "Hadir"
		if lateMin > 0 {
			status = "Terlambat"
		}
		if earlyMin > 0 && overtime == 0 {
			status = "Pulang Dulu"
		}
		if overtime > 0 {
			status = "Lembur"
		}

		reports = append(reports, DailyReport{
			Date:       key.date,
			EmployeeID: key.employeeID,
			DeviceSN:   key.deviceSN,
			CheckIn:    checkIn.Format("15:04:05"),
			CheckOut:   checkOut.Format("15:04:05"),
			LateMin:    lateMin,
			EarlyMin:   earlyMin,
			Overtime:   overtime,
			Status:     status,
		})
	}

	// Sort reports by date DESC, employee_id ASC.
	sort.Slice(reports, func(i, j int) bool {
		if reports[i].Date != reports[j].Date {
			return reports[i].Date > reports[j].Date
		}
		return reports[i].EmployeeID < reports[j].EmployeeID
	})

	// Summary.
	summary := ReportSummary{
		TotalRecords: len(reports),
		DateFrom:     filter.DateFrom.Format("2006-01-02"),
		DateTo:       filter.DateTo.Format("2006-01-02"),
	}
	for _, r := range reports {
		switch r.Status {
		case "Hadir":
			summary.Hadir++
		case "Terlambat":
			summary.Terlambat++
		}
	}

	return &ReportResult{
		Reports: reports,
		Summary: summary,
	}, nil
}

// GenerateCSV returns the report as a CSV string.
func (s *ReportService) GenerateCSV(ctx context.Context, filter ReportFilter) (string, error) {
	result, err := s.GenerateAttendanceReport(ctx, filter)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("Date,Employee ID,Device,Check In,Check Out,Late (min),Early (min),Overtime (min),Status\n")
	for _, r := range result.Reports {
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%d,%d,%d,%s\n",
			r.Date, r.EmployeeID, r.DeviceSN,
			r.CheckIn, r.CheckOut,
			r.LateMin, r.EarlyMin, r.Overtime, r.Status))
	}
	return sb.String(), nil
}

func parseTime(t string) time.Time {
	now := time.Now()
	parsed, err := time.ParseInLocation("15:04", t, now.Location())
	if err != nil {
		return time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
	}
	return time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())
}
