package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
	"zadanie-6105/internal/Storage/postgresql"
	"zadanie-6105/internal/config"
	ping "zadanie-6105/internal/http-server/handlers"
	"zadanie-6105/internal/http-server/handlers/bids/addFeedback"
	createBid "zadanie-6105/internal/http-server/handlers/bids/create"
	editBid "zadanie-6105/internal/http-server/handlers/bids/edit"
	editBidStatus "zadanie-6105/internal/http-server/handlers/bids/editStatus"
	bidList "zadanie-6105/internal/http-server/handlers/bids/getList"
	personalBid "zadanie-6105/internal/http-server/handlers/bids/getPersonal"
	"zadanie-6105/internal/http-server/handlers/bids/getReviews"
	bidRollback "zadanie-6105/internal/http-server/handlers/bids/getRollback"
	bidStatus "zadanie-6105/internal/http-server/handlers/bids/getStatus"
	"zadanie-6105/internal/http-server/handlers/bids/submitDecision"
	"zadanie-6105/internal/http-server/handlers/tender/checkStatus"
	"zadanie-6105/internal/http-server/handlers/tender/create"
	"zadanie-6105/internal/http-server/handlers/tender/edit"
	"zadanie-6105/internal/http-server/handlers/tender/editStatus"
	"zadanie-6105/internal/http-server/handlers/tender/getAll"
	"zadanie-6105/internal/http-server/handlers/tender/getPersonal"
	tenderRollback "zadanie-6105/internal/http-server/handlers/tender/getRollback"
	mwLogger "zadanie-6105/internal/http-server/middleware/logger"
)

const (
	envLocal = "local"
	envDev   = "dev"
)

func main() {

	conf := config.NewConfig()

	logger := setupLogger(conf.Env)

	logger.Info("Starting AvitoTender service", slog.String("env", conf.Env))
	logger.Debug("debug msg are enabled")

	storage, err := postgresql.New(conf.StoragePath)

	if err != nil {
		logger.Error(err.Error(), slog.String("error", err.Error()))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/api", func(r chi.Router) {
		r.Get("/ping", ping.New(logger))
		r.Route("/tenders", func(r chi.Router) {
			r.Get("/", getAll.New(logger, storage))
			r.Post("/new", create.New(logger, storage))
			r.Get("/my", getPersonal.New(logger, storage))
			r.Route("/{tenderId}", func(r chi.Router) {
				r.Get("/status", checkStatus.New(logger, storage))
				r.Put("/status", editStatus.New(logger, storage))
				r.Patch("/edit", edit.New(logger, storage))
				r.Put("/rollback/{version}", tenderRollback.New(logger, storage))
			})
		})

		r.Route("/bids", func(r chi.Router) {
			r.Post("/new", createBid.New(logger, storage))
			r.Get("/my", personalBid.New(logger, storage))
			r.Get("/{tenderId}/list", bidList.New(logger, storage))
			r.Get("/{tenderId}/reviews", getReviews.New(logger, storage))
			r.Route("/{bidId}", func(r chi.Router) {
				r.Get("/status", bidStatus.New(logger, storage))
				r.Put("/status", editBidStatus.New(logger, storage))
				r.Patch("/edit", editBid.New(logger, storage))
				r.Put("/submit_decision", submitDecision.New(logger, storage))
				r.Put("/feedback", addFeedback.New(logger, storage))
				r.Put("/rollback/{version}", bidRollback.New(logger, storage))
			})
		})
	})

	logger.Info("starting server", slog.String("address", conf.Address))

	srv := &http.Server{
		Addr:         conf.Address,
		Handler:      router,
		ReadTimeout:  conf.Timeout,
		WriteTimeout: conf.Timeout,
		IdleTimeout:  conf.IdleTimeout,
	}
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("failed to start server")
	}

	logger.Error("server stopped")
}

func setupLogger(env string) *slog.Logger {

	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	}

	return log
}
