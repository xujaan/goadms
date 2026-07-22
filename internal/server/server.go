package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jan/goadms/internal/config"
	"github.com/jan/goadms/internal/handler"
	mymw "github.com/jan/goadms/internal/middleware"
	"github.com/jan/goadms/internal/repository"
	"github.com/jan/goadms/internal/service"
	"github.com/jan/goadms/internal/webhook"
)

type Server struct {
	cfg    *config.Config
	router *chi.Mux
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func New(cfg *config.Config, logger *slog.Logger) (*Server, error) {
	pool, err := pgxpool.New(context.Background(), cfg.Database.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	if cfg.Database.PoolMax > 0 {
		pool.Config().MaxConns = int32(cfg.Database.PoolMax)
	}

	s := &Server{
		cfg:    cfg,
		router: chi.NewRouter(),
		pool:   pool,
		logger: logger,
	}

	s.setup()
	return s, nil
}

func (s *Server) setup() {
	// --- Repositories ---
	userRepo := repository.NewAppUserRepo(s.pool)
	deviceRepo := repository.NewDeviceRepo(s.pool)
	attendanceRepo := repository.NewAttendanceRepo(s.pool)
	rawLogRepo := repository.NewRawLogRepo(s.pool)
	refreshRepo := repository.NewRefreshTokenRepo(s.pool)
	webhookRepo := repository.NewWebhookRepo(s.pool)

	// --- Webhook System ---
	signer := webhook.NewSigner()
	deliverer := webhook.NewDeliverer(
		webhookRepo, signer,
		s.cfg.Webhook.TimeoutDuration(),
		s.cfg.Webhook.RetryMax(),
		s.cfg.Webhook.RetryBaseDuration(),
		s.logger,
	)
	dispatcher := webhook.NewDispatcher(webhookRepo, deliverer, s.logger)

	// --- SSE Broker ---
	sseBroker := service.NewSSEBroker(s.logger)

	// --- Services ---
	authSvc := service.NewAuthService(userRepo, refreshRepo, s.cfg.Auth)
	deviceSvc := service.NewDeviceService(deviceRepo, rawLogRepo, dispatcher)
	pushSvc := service.NewPushService(attendanceRepo, rawLogRepo, deviceSvc, dispatcher)
	webhookSvc := service.NewWebhookService(webhookRepo)
	zktecoSvc := service.NewZkTecoService(deviceRepo, attendanceRepo, dispatcher, s.logger)
	shiftSvc := service.NewShiftService(repository.NewShiftRepo(s.pool))
	reportSvc := service.NewReportService(attendanceRepo, repository.NewShiftRepo(s.pool), deviceRepo)
	fingerSvc := service.NewFingerUserService(repository.NewFingerUserRepo(s.pool), deviceRepo, s.logger)
	scannerSvc := service.NewScannerService(s.logger)

	// Wire SSE into push service (broadcast on attendance created).
	pushSvc.SetSSEBroker(sseBroker)

	// --- Handlers ---
	webH := handler.NewWebHandler(deviceRepo, attendanceRepo, "templates")
	authH := handler.NewAuthHandler(authSvc)
	deviceH := handler.NewDeviceHandler(deviceSvc, zktecoSvc)
	pushH := handler.NewPushHandler(deviceSvc, pushSvc, rawLogRepo)
	webhookH := handler.NewWebhookHandler(webhookSvc)
	shiftH := handler.NewShiftHandler(shiftSvc)
	reportH := handler.NewReportHandler(reportSvc)
	attendanceH := handler.NewAttendanceHandler(attendanceRepo)
	fingerH := handler.NewFingerUserHandler(fingerSvc)
	sseH := handler.NewSSEHandler(sseBroker)
	scannerH := handler.NewScannerHandler(scannerSvc)

	// --- Middleware ---
	s.router.Use(chimw.RequestID)
	s.router.Use(chimw.RealIP)
	s.router.Use(mymw.Logger(s.logger))
	s.router.Use(mymw.CORS)
	s.router.Use(chimw.Recoverer)

	// --- Public routes ---
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Favicon — silent 204.
	s.router.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })

	// Root redirect to dashboard.
	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/dashboard", http.StatusFound)
	})

	// Static files.
	s.router.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	// Web UI routes (no auth — JS handles token in browser).
	s.router.Get("/web/login", webH.LoginPage)
	s.router.Get("/web/dashboard", webH.Dashboard)
	s.router.Get("/web/devices", webH.DevicesPage)
	s.router.Get("/web/attendance", webH.AttendancePage)
	s.router.Get("/web/shifts", webH.ShiftsPage)
	s.router.Get("/web/reports", webH.ReportsPage)
	s.router.Get("/web/fingerprint-users", webH.FingerUsersPage)
	s.router.Get("/web/scanner", webH.ScannerPage)
	s.router.Get("/web/webhooks", webH.WebhooksPage)
		s.router.Get("/web/device-users", webH.DeviceUsersPage)
	// SSE events (no auth — JS EventSource can't set headers).
	s.router.Get("/api/v1/events", sseH.Stream)

	// ZKTeco push protocol (no auth).
	s.router.Get("/iclock/cdata", pushH.Handshake)
	s.router.Post("/iclock/cdata", pushH.ReceiveRecords)
	s.router.Get("/iclock/test", pushH.Test)
	s.router.Get("/iclock/getrequest", pushH.GetRequest)

	// Auth routes (no auth required).
	s.router.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/login", authH.Login)
		r.Post("/refresh", authH.Refresh)
		r.With(mymw.AuthMiddleware(s.cfg.Auth.JWTSecret)).Post("/logout", authH.Logout)
		r.With(mymw.AuthMiddleware(s.cfg.Auth.JWTSecret)).Get("/me", authH.Me)
	})

	// Protected API routes.
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Use(mymw.AuthMiddleware(s.cfg.Auth.JWTSecret))

		// Devices.
		r.Get("/devices", deviceH.List)
		r.Post("/devices", deviceH.Create) // TODO: add AdminOnly middleware for creation
		r.Get("/devices/{id}", deviceH.GetByID)
		r.Put("/devices/{id}", deviceH.Update)
		r.Delete("/devices/{id}", deviceH.Delete)

			// TCP device operations.
			r.Post("/devices/{id}/test-connection", deviceH.TestConnection)
			r.Post("/devices/{id}/pull-attendance", deviceH.PullAttendance)
			r.Get("/devices/{id}/users", deviceH.ListDeviceUsers)
			r.Post("/devices/{id}/reboot", deviceH.Reboot)
			r.Post("/devices/{id}/sync-time", deviceH.SyncTime)
			r.Post("/devices/{id}/users/{uid}/delete", deviceH.DeleteDeviceUser)

			// Shifts.
			r.Get("/shifts", shiftH.List)
			r.Post("/shifts", shiftH.Create)
			r.Put("/shifts/{id}", shiftH.Update)
			r.Delete("/shifts/{id}", shiftH.Delete)
			r.Post("/shifts/{id}/assign", shiftH.AssignUser)
			r.Get("/shifts/{id}/users", shiftH.GetAssignedUsers)

			// Fingerprint users.
			r.Get("/fingerprint-users", fingerH.List)
			r.Post("/fingerprint-users", fingerH.Create)
			r.Put("/fingerprint-users/{id}", fingerH.Update)
			r.Delete("/fingerprint-users/{id}", fingerH.Delete)
			r.Post("/fingerprint-users/{id}/sync", fingerH.SyncToDevice)
			r.Post("/fingerprint-users/sync-all", fingerH.SyncAllToDevice)

			// Network scanner.
			r.Post("/detect/scan", scannerH.ScanSubnet)
			r.Post("/detect/one", scannerH.DetectSingle)

			// Attendance logs.
			r.Get("/attendances", attendanceH.List)
			r.Delete("/attendances/{id}", attendanceH.Delete)

			// Reports.
			r.Get("/reports/attendance", reportH.AttendanceReport)
			r.Get("/reports/attendance.csv", reportH.AttendanceCSV)

			// Webhooks per device.
		r.Get("/devices/{id}/webhooks", webhookH.List)
		r.Post("/devices/{id}/webhooks", webhookH.Create)
		r.Put("/devices/{id}/webhooks/{wid}", webhookH.Update)
		r.Delete("/devices/{id}/webhooks/{wid}", webhookH.Delete)
		r.Post("/devices/{id}/webhooks/{wid}/test", webhookH.TestPing)
		r.Get("/devices/{id}/webhooks/{wid}/deliveries", webhookH.ListDeliveries)
	})
}

// Run starts the HTTP server and handles graceful shutdown.
func (s *Server) Run() error {
	addr := s.cfg.Server.Address()
	srv := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  s.cfg.Server.ReadTimeoutDuration(),
		WriteTimeout: s.cfg.Server.WriteTimeoutDuration(),
		IdleTimeout:  s.cfg.Server.IdleTimeoutDuration(),
	}

	// Run scheduler goroutines.
	s.runSchedulers()

	// Graceful shutdown.
	idleConnsClosed := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig

		s.logger.Info("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			s.logger.Error("shutdown error", "error", err)
		}
		s.pool.Close()
		close(idleConnsClosed)
	}()

	s.logger.Info("server starting", "addr", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	<-idleConnsClosed
	s.logger.Info("server stopped")
	return nil
}

func (s *Server) runSchedulers() {
	// Device online checker: every 60s.
	go func() {
		interval := s.cfg.Scheduler.OnlineCheckDuration()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			deviceRepo := repository.NewDeviceRepo(s.pool)
			offline, err := deviceRepo.ListOffline(context.Background(), 5*time.Minute)
			if err != nil {
				s.logger.Error("online checker", "error", err)
				continue
			}
			s.logger.Debug("online checker", "offline_count", len(offline))
		}
	}()

	// Auto-pull attendance from all devices: configurable interval.
	go func() {
		interval := s.cfg.Scheduler.AutoPullDuration()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			deviceRepo := repository.NewDeviceRepo(s.pool)
			attendanceRepo := repository.NewAttendanceRepo(s.pool)
			webhookRepo := repository.NewWebhookRepo(s.pool)
			signer_ := webhook.NewSigner()
			deliverer_ := webhook.NewDeliverer(
				webhookRepo, signer_,
				s.cfg.Webhook.TimeoutDuration(),
				s.cfg.Webhook.RetryMax(),
				s.cfg.Webhook.RetryBaseDuration(),
				s.logger,
			)
			dispatcher_ := webhook.NewDispatcher(webhookRepo, deliverer_, s.logger)
			zkSvc := service.NewZkTecoService(deviceRepo, attendanceRepo, dispatcher_, s.logger)
			zkSvc.AutoPullAll(context.Background())
		}
	}()

	s.logger.Info("schedulers started",
		"online_check_interval", s.cfg.Scheduler.OnlineCheckDuration(),
		"auto_pull_interval", s.cfg.Scheduler.AutoPullDuration(),
	)
}
