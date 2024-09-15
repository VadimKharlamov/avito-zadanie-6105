package editStatus

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	tenderOpt "zadanie-6105/internal/lib/api/DTO"
	"zadanie-6105/internal/lib/api/response"
)

type StatusEditor interface {
	EditTenderStatus(tenderId, status, username string) (tenderOpt.TenderResponse, error)
	GetUserId(username string) (string, error)
	ValidateRules(userId, organizationId string) (bool, error)
	GetTenderOrganizationId(tenderId string) (string, error)
}

func New(log *slog.Logger, statusEditor StatusEditor) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.tender.editStatus"

		tenderId := chi.URLParam(request, "tenderId")
		if tenderId == "" {
			log.Error("no tenderId provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no tenderId provided"))
			return
		}

		username := request.URL.Query().Get("username")
		if username == "" {
			log.Error("no username provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no username provided"))
			return
		}

		status := request.URL.Query().Get("status")
		if status == "" || !(status == "Created" || status == "Published" || status == "Closed") {
			log.Error("no new status provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no new correct status provided"))
			return
		}

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		id, err := statusEditor.GetUserId(username)
		if err != nil {
			log.Error("failed to get user", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("incorrect username or not provided"))
			return
		}

		organizationId, err := statusEditor.GetTenderOrganizationId(tenderId)
		if err != nil {
			log.Error("failed to get organizationId", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("no tender with this id"))
			return
		}

		rules, err := statusEditor.ValidateRules(id, organizationId)
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

		resp, err := statusEditor.EditTenderStatus(tenderId, status, username)
		if err != nil {
			log.Error("failed to edit tender status", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			render.JSON(writer, request, response.Error("fail to edit tender status"))
			return
		}

		log.Info("success to edit tender status", slog.Any("response", resp))
		render.JSON(writer, request, resp)
	}
}
