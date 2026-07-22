package service

import (
	"context"
	"fmt"

	"github.com/jan/goadms/internal/model"
	"github.com/jan/goadms/internal/repository"
	"github.com/jan/goadms/internal/webhook"
)

type PushService struct {
	attendanceRepo *repository.AttendanceRepo
	rawLogRepo     *repository.RawLogRepo
	deviceSvc      *DeviceService
	dispatcher     *webhook.Dispatcher
	sseBroker      *SSEBroker
}

func NewPushService(attendanceRepo *repository.AttendanceRepo, rawLogRepo *repository.RawLogRepo, deviceSvc *DeviceService, dispatcher *webhook.Dispatcher) *PushService {
	return &PushService{
		attendanceRepo: attendanceRepo,
		rawLogRepo:     rawLogRepo,
		deviceSvc:      deviceSvc,
		dispatcher:     dispatcher,
	}
}

func (s *PushService) SetSSEBroker(broker *SSEBroker) {
	s.sseBroker = broker
}

// ProcessRecords parses tab-separated body, deduplicates, inserts, and fires webhooks.
func (s *PushService) ProcessRecords(ctx context.Context, sn, body, table string) (int, error) {
	records := ParseBodyAsTabSeparated(body)
	if len(records) == 0 {
		return 0, nil
	}

	for i := range records {
		records[i].DeviceSN = sn
	}

	source := "push"
	sourceFromTable := tableToSource(table)
	if sourceFromTable != "" {
		source = sourceFromTable
	}

	inserted, err := s.attendanceRepo.BulkInsert(ctx, records, source)
	if err != nil {
		return 0, fmt.Errorf("bulk insert attendance: %w", err)
	}

	// Log raw request.
	s.rawLogRepo.Insert(ctx, &model.RawLog{
		DeviceSN:      sn,
		RequestMethod: "POST",
		RequestURI:    "/iclock/cdata",
		RequestBody:   body,
		LogType:       "push_records",
	})

	// Broadcast to SSE clients.
	if s.sseBroker != nil {
		for _, rec := range records {
			s.sseBroker.Broadcast("attendance", map[string]any{
				"device_sn":   sn,
				"employee_id": rec.EmployeeID,
				"timestamp":   rec.Timestamp,
				"source":      source,
			})
		}
	}

	// Fire webhooks for each inserted record.
	device, _ := s.deviceSvc.GetBySN(ctx, sn)
	if device != nil {
		for _, rec := range records {
			s.dispatcher.FanOut(ctx, "attendance.created", device.ID, sn, map[string]any{
				"device_sn":   sn,
				"employee_id": rec.EmployeeID,
				"timestamp":   rec.Timestamp,
				"status1":     rec.Status1,
				"status2":     rec.Status2,
				"status3":     rec.Status3,
				"status4":     rec.Status4,
				"status5":     rec.Status5,
				"source":      source,
			})
		}
	}

	return inserted, nil
}

func tableToSource(table string) string {
	switch table {
	case "ATTLOG":
		return "push"
	case "OPERLOG":
		return "push"
	default:
		return "push"
	}
}
