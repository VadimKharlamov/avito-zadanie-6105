package edit

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	tenderOpt "zadanie-6105/internal/lib/api/DTO"
	"zadanie-6105/internal/lib/api/response"
)

type TenderEditor interface {
	EditTender(req tenderOpt.TenderPatchRequest, tenderId, username string) (tenderOpt.TenderResponse, error)
	ValidateRules(userId, organizationId string) (bool, error)
	GetUserId(username string) (string, error)
	GetTenderOrganizationId(tenderId string) (string, error)
}

func New(log *slog.Logger, tenderEditor TenderEditor) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.tender.editTender"

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

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		var req tenderOpt.TenderPatchRequest

		err := render.DecodeJSON(request.Body, &req)
		if err != nil {
			log.Error("failed to deserialize request", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("fail to decode request"))
			return
		}

		log.Info("request tender body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			var validationErrors validator.ValidationErrors
			errors.As(err, &validationErrors)
			log.Error("fail to validate request", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.ValidationError(validationErrors))

			return
		}

		id, err := tenderEditor.GetUserId(username)
		if err != nil {
			log.Error("failed to get user", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("no user with this username"))
			return
		}

		organizationId, err := tenderEditor.GetTenderOrganizationId(tenderId)
		if err != nil {
			log.Error("failed to get organizationId", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("no tender with this id"))
			return
		}

		rules, err := tenderEditor.ValidateRules(id, organizationId)
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

		resp, err := tenderEditor.EditTender(req, tenderId, username)

		if err != nil {
			log.Error("failed to edit tender", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			render.JSON(writer, request, response.Error("fail to edit tender"))
			return
		}

		log.Info("success to edit tender", slog.Any("response", resp))
		render.JSON(writer, request, resp)
	}
}
