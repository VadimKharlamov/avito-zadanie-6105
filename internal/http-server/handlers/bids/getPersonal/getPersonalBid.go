package getPersonal

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	bidOpt "zadanie-6105/internal/lib/api/DTO"
	"zadanie-6105/internal/lib/api/response"
)

type BidGetter interface {
	GetPersonalBids(limit string, offset string, id string) ([]bidOpt.BidResponse, error)
	GetUserId(username string) (string, error)
}

func New(log *slog.Logger, bidGetter BidGetter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.tender.getPersonalBid"

		limit := request.URL.Query().Get("limit")
		if limit == "" {
			limit = "5"
		}

		offset := request.URL.Query().Get("offset")
		if offset == "" {
			offset = "0"
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

		id, err := bidGetter.GetUserId(username)
		if err != nil {
			log.Error("user not found", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("user not found"))
			return
		}

		resp, err := bidGetter.GetPersonalBids(limit, offset, id)
		if err != nil {
			log.Error("failed to get bids", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			render.JSON(writer, request, response.Error("fail to get bids"))
			return
		}

		log.Info("success to get bids", slog.Any("response", resp))
		render.JSON(writer, request, resp)
	}
}
