package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/ilham/presensi-online/backend/internal/config"
	"github.com/ilham/presensi-online/backend/internal/repository/database"
	"github.com/ilham/presensi-online/backend/internal/repository/postgres"
	"github.com/ilham/presensi-online/backend/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	if cfg.App.Env == "production" {
		logger, _ = zap.NewProduction()
	}
	defer logger.Sync()

	ctx := context.Background()

	pgPool, err := database.NewPostgresPool(ctx, &cfg.Database)
	if err != nil {
		logger.Fatal("postgres connection failed", zap.Error(err))
	}
	defer pgPool.Close()

	repos := worker.Repos{
		Faculty:      postgres.NewFacultyRepo(pgPool),
		StudyProgram: postgres.NewStudyProgramRepo(pgPool),
		Building:     postgres.NewBuildingRepo(pgPool),
		Room:         postgres.NewRoomRepo(pgPool),
		User:         postgres.NewUserRepo(pgPool),
		Schedule:     postgres.NewClassScheduleRepo(pgPool),
		StudyPlan:    postgres.NewStudyPlanRepo(pgPool),
	}

	syncWorker, err := worker.NewSyncWorker(cfg, repos, logger)
	if err != nil {
		logger.Fatal("failed to create sync worker", zap.Error(err))
	}

	// Run immediately on startup
	logger.Info("sync worker: initial run on startup")
	if err := syncWorker.Run(ctx); err != nil {
		logger.Error("initial sync failed", zap.Error(err))
	}

	// Schedule periodic sync via cron
	c := cron.New(cron.WithLocation(time.Local))
	_, err = c.AddFunc(cfg.Worker.Cron, func() {
		logger.Info("sync worker: cron triggered")
		if err := syncWorker.Run(context.Background()); err != nil {
			logger.Error("scheduled sync failed", zap.Error(err))
		}
	})
	if err != nil {
		logger.Fatal("failed to schedule cron", zap.Error(err))
	}

	c.Start()
	logger.Info("sync worker started", zap.String("cron", cfg.Worker.Cron))

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("sync worker shutting down...")
	c.Stop()
	logger.Info("sync worker stopped")
}
