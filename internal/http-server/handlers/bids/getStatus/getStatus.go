package getStatus

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"zadanie-6105/internal/lib/api/response"
)

type BidGetter interface {
	GetBidStatus(bidId string) (string, error)
	GetUserId(username string) (string, error)
	IsHavePerms(userId, bidId string) (bool, error)
}

func New(log *slog.Logger, bidGetter BidGetter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.tender.getBidStatus"

		bidId := chi.URLParam(request, "bidId")
		if bidId == "" {
			log.Error("no bidId provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no bidId provided"))
			return
		}

		username := request.URL.Query().Get("username")
		if username == "" {
			log.Error("no username provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no username provided"))
			return
		}

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		userId, err := bidGetter.GetUserId(username)
		if err != nil {
			log.Error("no user found")
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("no user with this nickname"))
			return
		}

		perms, err := bidGetter.IsHavePerms(userId, bidId)
		if err != nil || !perms {
			log.Error("no permission")
			writer.WriteHeader(http.StatusForbidden)
			render.JSON(writer, request, response.Error("no permission"))
			return
		}

		resp, err := bidGetter.GetBidStatus(bidId)
		if err != nil {
			log.Error("failed to get bid status", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			render.JSON(writer, request, response.Error("fail to get bid status"))
			return
		}

		log.Info("success to get bid status", slog.Any("response", resp))
		render.JSON(writer, request, resp)
	}
}
