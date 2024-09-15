package checkStatus

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"zadanie-6105/internal/lib/api/response"
)

type TenderGetter interface {
	GetTenderStatus(tenderId string) (string, error)
	GetUserId(username string) (string, error)
	GetTenderOrganizationId(tenderId string) (string, error)
	ValidateRules(userId, organizationId string) (bool, error)
	IsTenderPublic(tenderId string) (bool, error)
}

func New(log *slog.Logger, tenderGetter TenderGetter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.tender.getTenderStatus"

		tenderId := chi.URLParam(request, "tenderId")
		if tenderId == "" {
			log.Error("no tenderId provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no tenderId provided"))
			return
		}

		username := request.URL.Query().Get("username")

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		isPublic, err := tenderGetter.IsTenderPublic(tenderId)
		if err != nil {
			log.Error("failed to get tender status", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("no tender with this id"))
			return
		}

		if !isPublic {
			id, err := tenderGetter.GetUserId(username)
			if err != nil {
				log.Error("failed to get user", slog.String("error", err.Error()))
				writer.WriteHeader(http.StatusBadRequest)
				render.JSON(writer, request, response.Error("incorrect username or not provided"))
				return
			}

			organizationId, err := tenderGetter.GetTenderOrganizationId(tenderId)
			if err != nil {
				log.Error("failed to get organizationId", slog.String("error", err.Error()))
				writer.WriteHeader(http.StatusNotFound)
				render.JSON(writer, request, response.Error("no tender with this id"))
				return
			}

			rules, err := tenderGetter.ValidateRules(id, organizationId)
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
		}

		resp, err := tenderGetter.GetTenderStatus(tenderId)
		if err != nil {
			log.Error("failed to get tender status", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			render.JSON(writer, request, response.Error("fail to get tender status"))
			return
		}

		log.Info("success to get tender status", slog.Any("response", resp))
		render.JSON(writer, request, resp)
	}
}
