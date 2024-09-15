package create

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	tenderOpt "zadanie-6105/internal/lib/api/DTO"
	"zadanie-6105/internal/lib/api/response"
)

type TenderSaver interface {
	CreateTender(tenderData tenderOpt.TenderRequest) (tenderOpt.TenderResponse, error)
	GetUserId(username string) (string, error)
	ValidateRules(userId, organizationId string) (bool, error)
}

func New(log *slog.Logger, tenderSaver TenderSaver) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.tender.create"

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		var req tenderOpt.TenderRequest

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

		id, err := tenderSaver.GetUserId(req.CreatorUsername)
		if err != nil {
			log.Error("failed to get user", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("no user with this username"))
			return
		}

		rules, err := tenderSaver.ValidateRules(id, req.OrganizationId)
		if err != nil {
			log.Error("failed to check permission", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			render.JSON(writer, request, response.Error("failed to check permission"))
			return
		}
		if !rules {
			log.Error("no permission", slog.String("error", "no permission"))
			writer.WriteHeader(http.StatusForbidden)
			render.JSON(writer, request, response.Error("no permission to create tender"))
			return
		}
		resp, err := tenderSaver.CreateTender(req)
		if err != nil {
			log.Error("failed to save tender", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			render.JSON(writer, request, response.Error("fail to save tender"))
			return
		}

		log.Info("success to create tender", slog.Any("response", resp))
		render.JSON(writer, request, resp)
	}
}
