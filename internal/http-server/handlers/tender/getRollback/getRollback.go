package getRollback

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	tenderOpt "zadanie-6105/internal/lib/api/DTO"
	"zadanie-6105/internal/lib/api/response"
)

type RollbackGetter interface {
	GetTenderRollback(tenderId, version, username string) (tenderOpt.TenderResponse, error)
	ValidateRules(userId, organizationId string) (bool, error)
	GetUserId(username string) (string, error)
	GetTenderOrganizationId(tenderId string) (string, error)
}

func New(log *slog.Logger, rollbackGetter RollbackGetter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.tender.getRollback"

		tenderId := chi.URLParam(request, "tenderId")
		if tenderId == "" {
			log.Error("no tenderId provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no tenderId provided"))
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

		organizationId, err := rollbackGetter.GetTenderOrganizationId(tenderId)
		if err != nil {
			log.Error("failed to get organizationId", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("no tender with this id"))
			return
		}

		rules, err := rollbackGetter.ValidateRules(id, organizationId)
		if err != nil {
			log.Error("failed to check permission", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			render.JSON(writer, request, response.Error("failed to check permission"))
			return
		}
		if !rules {
			log.Error("no permission", slog.String("error", "no permission"))
			writer.WriteHeader(http.StatusForbidden)
			render.JSON(writer, request, response.Error("no permission to get tender status"))
			return
		}

		resp, err := rollbackGetter.GetTenderRollback(tenderId, version, username)

		if err != nil {
			log.Error("failed to rollback tender", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			render.JSON(writer, request, response.Error("fail to rollback tender"))
			return
		}

		log.Info("success to rollback tender", slog.Any("response", resp))
		render.JSON(writer, request, resp)
	}
}
