package getPersonal

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"sort"
	tenderOpt "zadanie-6105/internal/lib/api/DTO"
	"zadanie-6105/internal/lib/api/response"
)

type TenderGetter interface {
	GetPersonalTenders(limit string, offset string, username string) ([]tenderOpt.TenderResponse, error)
}

func New(log *slog.Logger, tenderGetter TenderGetter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.tender.getPersonalTender"

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

		resp, err := tenderGetter.GetPersonalTenders(limit, offset, username)
		if err != nil {
			log.Error("failed to get tenders", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("fail to get tenders"))
			return
		}

		sort.Slice(resp, func(i, j int) bool {
			return resp[i].Name < resp[j].Name
		})

		log.Info("success to get tenders", slog.Any("response", resp))
		render.JSON(writer, request, resp)
	}
}
