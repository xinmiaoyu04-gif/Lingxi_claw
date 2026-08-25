// Command server starts the Lingxi-claw HTTP API described in API.md.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lingxi-claw/internal/agent"
	"lingxi-claw/internal/config"
	"lingxi-claw/internal/handler"
	"lingxi-claw/internal/repository"
	"lingxi-claw/internal/service"
	"lingxi-claw/internal/workflow"
)

func main() {
	cfg := config.Load()
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           buildHandler(cfg, log),
		ReadHeaderTimeout: 10 * time.Second,
		// Uploads can be large and OCR-bound, so the write timeout is generous.
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Shut down cleanly on Ctrl-C / SIGTERM so in-flight requests finish.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Info("server listening", "addr", cfg.Addr, "mode", cfg.Mode)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		log.Error("server failed", "error", err)
		os.Exit(1)
	case <-ctx.Done():
		log.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}

// buildHandler wires repository → service → workflow → handler, matching the
// layering in API.md §15.
func buildHandler(cfg config.Config, log *slog.Logger) http.Handler {
	store := repository.New()

	// In mock mode OCR is a fixture backend (API.md §14). Real mode swaps this
	// implementation only; every layer above keeps the same response fields.
	parser := service.NewParser(service.NewMockOCR())
	questions := service.NewQuestionService(store)
	analysis := service.NewAnalysisService()
	plans := service.NewPlanService()
	grading := service.NewGradingService()
	hints := service.NewHintService()

	finalSprint := workflow.NewFinalSprint(store, parser, questions, analysis, plans, grading, log)
	homework := workflow.NewHomework(store, parser, questions, hints, grading, log)
	general := agent.NewGeneral()

	h := handler.New(cfg, store, finalSprint, homework, general, log)

	return handler.Chain(h.Routes(),
		handler.Recover(log),
		handler.AccessLog(log),
		handler.CORS(cfg.AllowedOrigins),
		handler.NotFoundEnvelope,
	)
}
