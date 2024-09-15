package ping

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

func New(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const op = "handlers.ping.New"

		log = log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(req.Context())))

		log.Info("pinged server", "status", "done")
		render.JSON(w, req, "Ok")
	}
}
