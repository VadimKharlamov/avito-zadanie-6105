package getList

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	bidOpt "zadanie-6105/internal/lib/api/DTO"
	"zadanie-6105/internal/lib/api/response"
)

type BidGetter interface {
	GetBidList(limit, offset, tenderId string) ([]bidOpt.BidResponse, error)
	ValidateRules(userId, organizationId string) (bool, error)
	GetUserId(username string) (string, error)
	GetTenderOrganizationId(tenderId string) (string, error)
}

func New(log *slog.Logger, bidGetter BidGetter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.tender.getBidList"

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

		tenderId := chi.URLParam(request, "tenderId")
		if tenderId == "" {
			log.Error("no tenderId provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no tenderId provided"))
			return
		}

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		id, err := bidGetter.GetUserId(username)
		if err != nil {
			log.Error("failed to get user", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("no user with this username"))
			return
		}

		organizationId, err := bidGetter.GetTenderOrganizationId(tenderId)
		if err != nil {
			log.Error("failed to get organizationId", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("no tender with this id"))
			return
		}

		rules, err := bidGetter.ValidateRules(id, organizationId)
		if err != nil {
			log.Error("failed to check permission", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			render.JSON(writer, request, response.Error("failed to check permission"))
			return
		}
		if !rules {
			log.Error("no permission", slog.String("error", "no permission"))
			writer.WriteHeader(http.StatusForbidden)
			render.JSON(writer, request, response.Error("no permission to get tender bids"))
			return
		}

		resp, err := bidGetter.GetBidList(limit, offset, tenderId)
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
