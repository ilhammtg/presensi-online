package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ilham/presensi-online/backend/internal/app"
	"github.com/ilham/presensi-online/backend/internal/config"
	"github.com/ilham/presensi-online/backend/internal/domain"
	"github.com/ilham/presensi-online/backend/internal/handler"
	authmw "github.com/ilham/presensi-online/backend/internal/middleware/auth"
	"github.com/ilham/presensi-online/backend/internal/repository/database"
	"github.com/ilham/presensi-online/backend/internal/repository/postgres"
	"github.com/ilham/presensi-online/backend/internal/repository/redisrepo"
	"github.com/ilham/presensi-online/backend/internal/ws"
)

func main() {
	// ─── Config ─────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// ─── Logger ─────────────────────────────────────────────────────────────
	var logger *zap.Logger
	if cfg.App.Env == "production" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	// ─── Database Connections ────────────────────────────────────────────────
	ctx := context.Background()

	pgPool, err := database.NewPostgresPool(ctx, &cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to PostgreSQL", zap.Error(err))
	}
	defer pgPool.Close()

	redisClient, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		logger.Fatal("failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	logger.Info("database connections established")

	// ─── Repositories ────────────────────────────────────────────────────────
	userRepo := postgres.NewUserRepo(pgPool)
	sessionRepo := postgres.NewClassSessionRepo(pgPool)
	attendanceRepo := postgres.NewAttendanceRepo(pgPool)
	scheduleRepo := postgres.NewClassScheduleRepo(pgPool)
	studyPlanRepo := postgres.NewStudyPlanRepo(pgPool)
	roomRepo := postgres.NewRoomRepo(pgPool)
	prodiRepo := postgres.NewProdiRepo(pgPool)
	systemRepo := postgres.NewSystemRepo(pgPool)

	qrCache := redisrepo.NewQRTokenCacheRepo(redisClient, cfg.QR.TOTPPeriodSeconds)
	sessionCache := redisrepo.NewSessionCacheRepo(redisClient)

	// ─── WebSocket Hub ────────────────────────────────────────────────────────
	hub := ws.NewHub()
	go hub.Run()

	// ─── Services & Use Cases ─────────────────────────────────────────────────
	jwtSvc := authmw.NewJWTService(cfg.JWT.Secret, cfg.JWT.AccessTTLMins, cfg.JWT.RefreshTTLDays)

	sessionUC := app.NewSessionUseCase(
		sessionRepo,
		scheduleRepo,
		studyPlanRepo,
		attendanceRepo,
		qrCache,
		sessionCache,
		hub,
		cfg.QR.TOTPPeriodSeconds,
	)

	attendanceUC := app.NewAttendanceUseCase(
		sessionRepo,
		studyPlanRepo,
		attendanceRepo,
		qrCache,
		scheduleRepo,
		roomRepo,
		systemRepo,
		hub,
	)

	// ─── Handlers ─────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(userRepo, jwtSvc)
	sessionHandler := handler.NewSessionHandler(sessionUC)
	attendanceHandler := handler.NewAttendanceHandler(attendanceUC)
	prodiHandler := handler.NewProdiHandler(prodiRepo, userRepo, systemRepo)
	systemHandler := handler.NewSystemHandler(systemRepo)
	wsHandler := handler.NewWSHandler(hub, jwtSvc)

	// ─── Router ───────────────────────────────────────────────────────────────
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware(cfg.CORS.AllowedOrigins))

	// Health check (no auth)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "presensi-api",
			"timestamp": time.Now().Unix(),
		})
	})

	// Direct WebSocket endpoint
	r.GET("/ws", wsHandler.ServeWS)

	v1 := r.Group("/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", authmw.Middleware(jwtSvc), authHandler.Me)
			auth.POST("/change-password", authmw.Middleware(jwtSvc), authHandler.ChangePassword)
			auth.POST("/profile-photo", authmw.Middleware(jwtSvc), authHandler.UpdateProfilePhoto)
		}

		// WebSocket endpoint under /v1
		v1.GET("/ws", wsHandler.ServeWS)

		// Protected routes
		protected := v1.Group("/")
		protected.Use(authmw.Middleware(jwtSvc))
		{
			// Dosen Schedules & Approval Center (Dual-Channel Permission)
			dosen := protected.Group("/dosen")
			dosen.Use(authmw.RequireRoles(domain.RoleDosen, domain.RoleAdminProdi, domain.RoleSuperadmin))
			{
				dosen.GET("/schedules", sessionHandler.GetDosenSchedules)
				dosen.GET("/permissions", sessionHandler.GetDosenPermissions)
				dosen.POST("/permissions/:id/approve", sessionHandler.ApproveDosenPermission)
			}

			// Mahasiswa Schedules (jadwal kuliah berdasarkan KRS enrollment)
			mahasiswa := protected.Group("/mahasiswa")
			mahasiswa.Use(authmw.RequireRoles(domain.RoleMahasiswa))
			{
				mahasiswa.GET("/schedules", sessionHandler.GetMahasiswaSchedules)
			}

			// Schedules Details (Students, Sessions history, Recap)
			schedules := protected.Group("/schedules")
			{
				schedules.GET("/:id/students", sessionHandler.GetScheduleStudents)
				schedules.GET("/:id/sessions", sessionHandler.GetScheduleSessions)
				schedules.GET("/:id/recap", sessionHandler.GetClassRecap)
			}

			// Sessions (Dosen & Admin)
			sessions := protected.Group("/sessions")
			{
				sessions.POST("", authmw.RequireRoles(domain.RoleDosen, domain.RoleAdminProdi, domain.RoleSuperadmin), sessionHandler.OpenSession)
				sessions.GET("/active", sessionHandler.GetActiveSession)
				sessions.PATCH("/:id/close", authmw.RequireRoles(domain.RoleDosen, domain.RoleAdminProdi, domain.RoleSuperadmin), sessionHandler.CloseSession)
				sessions.POST("/:id/permission", authmw.RequireRoles(domain.RoleDosen, domain.RoleAdminProdi, domain.RoleSuperadmin), sessionHandler.MarkPermission)
				sessions.GET("/:id", sessionHandler.GetSession)
				sessions.GET("/:id/attendees", authmw.RequireRoles(domain.RoleDosen, domain.RoleAdminProdi, domain.RoleSuperadmin), sessionHandler.GetAttendees)
				sessions.GET("/:id/qr", authmw.RequireRoles(domain.RoleDosen, domain.RoleAdminProdi, domain.RoleSuperadmin), sessionHandler.GetQR)
			}

			// Attendance Scanning (Mahasiswa — scan QR saja, izin diberikan langsung oleh dosen)
			attendance := protected.Group("/attendance")
			{
				attendance.POST("/scan", authmw.RequireRoles(domain.RoleMahasiswa), attendanceHandler.ScanAttendance)
			}

			// Attendance Management & Override (Dosen & Admin)
			attendances := protected.Group("/attendances")
			{
				attendances.PATCH("/:id", authmw.RequireRoles(domain.RoleDosen, domain.RoleAdminProdi), attendanceHandler.OverrideAttendance)
			}

			// Program Study Monitoring (Admin Prodi & Superadmin)
			prodi := protected.Group("/prodi")
			prodi.Use(authmw.RequireRoles(domain.RoleAdminProdi, domain.RoleSuperadmin))
			{
				prodi.GET("/overview", prodiHandler.GetProdiOverview)
				prodi.GET("/lecturers-compliance", prodiHandler.GetLecturersCompliance)
				prodi.GET("/students-at-risk", prodiHandler.GetStudentsAtRisk)
				prodi.GET("/classes", prodiHandler.GetProdiClasses)
				prodi.GET("/live-today", prodiHandler.GetLiveTodayClasses)
			}

			// System Administration (Superadmin only)
			system := protected.Group("/system")
			system.Use(authmw.RequireRoles(domain.RoleSuperadmin))
			{
				system.GET("/stats", systemHandler.GetSystemStats)
				system.GET("/study-programs", systemHandler.GetStudyPrograms)
				system.GET("/users", systemHandler.GetUsers)
				system.POST("/theme", systemHandler.SaveThemeSettings)
				system.GET("/config", systemHandler.GetCampusApiConfig)
				system.POST("/config", systemHandler.SaveCampusApiConfig)
				system.GET("/locations", systemHandler.GetAllCampusLocations)
				system.POST("/locations", systemHandler.CreateCampusLocation)
				system.PUT("/locations/:id", systemHandler.UpdateCampusLocation)
				system.DELETE("/locations/:id", systemHandler.DeleteCampusLocation)
				system.POST("/sync", systemHandler.TriggerSync)
			}
		}

		// Public Theme Settings and Active Locations endpoint
		v1.GET("/system/theme", systemHandler.GetThemeSettings)
		v1.GET("/system/locations/active", systemHandler.GetActiveCampusLocations)
	}

	// ─── HTTP Server with Graceful Shutdown ───────────────────────────────────
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("API server starting", zap.Int("port", cfg.App.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited gracefully")
}

// corsMiddleware adds CORS headers for the web dashboard.
func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	originSet := make(map[string]struct{})
	for _, o := range allowedOrigins {
		originSet[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, allowed := originSet[origin]; allowed || len(allowedOrigins) == 0 {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Device-ID")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
