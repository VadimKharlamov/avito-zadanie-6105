package getRollback

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	bidOpt "zadanie-6105/internal/lib/api/DTO"
	"zadanie-6105/internal/lib/api/response"
)

type RollbackGetter interface {
	GetBidRollback(bidId, version string) (bidOpt.BidResponse, error)
	GetUserId(username string) (string, error)
	IsHaveBidPerms(userId, bidId string) (bool, error)
}

func New(log *slog.Logger, rollbackGetter RollbackGetter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.bid.getRollback"

		bidId := chi.URLParam(request, "bidId")
		if bidId == "" {
			log.Error("no bidId provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no bidId provided"))
			return
		}

		version := chi.URLParam(request, "version")
		if version == "" {
			log.Error("no version provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no version provided"))
			return
		}

		username := request.URL.Query().Get("username")
		if version == "" {
			log.Error("no username provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no username provided"))
			return
		}

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		id, err := rollbackGetter.GetUserId(username)
		if err != nil {
			log.Error("failed to get user", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("no user with this username"))
			return
		}

		perms, err := rollbackGetter.IsHaveBidPerms(id, bidId)
		if err != nil {
			log.Error("bid not found", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("bid not found"))
			return
		}
		if !perms {
			log.Error("no permission")
			writer.WriteHeader(http.StatusForbidden)
			render.JSON(writer, request, response.Error("no permission"))
			return
		}

		resp, err := rollbackGetter.GetBidRollback(bidId, version)

		if err != nil {
			log.Error("failed to rollback bid", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			render.JSON(writer, request, response.Error("fail to rollback bid"))
			return
		}

		log.Info("success to rollback bid", slog.Any("response", resp))
		render.JSON(writer, request, resp)
	}
}
