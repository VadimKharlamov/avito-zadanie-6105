package getAll

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	tenderOpt "zadanie-6105/internal/lib/api/DTO"
	"zadanie-6105/internal/lib/api/response"
)

type TenderGetter interface {
	GetAllTenders(limit string, offset string, filters []string) ([]tenderOpt.TenderResponse, error)
}

func New(log *slog.Logger, tenderGetter TenderGetter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.tender.getAllTender"

		limit := request.URL.Query().Get("limit")
		if limit == "" {
			limit = "5"
		}

		offset := request.URL.Query().Get("offset")
		if offset == "" {
			offset = "0"
		}
		err := request.ParseForm()
		if err != nil {
			return
		}
		filters := request.Form["service_type"]

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		resp, err := tenderGetter.GetAllTenders(limit, offset, filters)
		if err != nil {
			log.Error("failed to get tenders", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("fail to save tender"))
			return
		}

		log.Info("success to save tender", slog.Any("response", resp))
		render.JSON(writer, request, resp)
	}
}
