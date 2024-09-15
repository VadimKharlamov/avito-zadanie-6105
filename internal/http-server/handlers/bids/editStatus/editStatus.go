package editStatus

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	bidOpt "zadanie-6105/internal/lib/api/DTO"
	"zadanie-6105/internal/lib/api/response"
)

type StatusEditor interface {
	EditBidStatus(bidId, status string) (bidOpt.BidResponse, error)
	IsHaveBidPerms(userId, bidId string) (bool, error)
	GetUserId(username string) (string, error)
}

func New(log *slog.Logger, statusEditor StatusEditor) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.bid.editStatus"

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

		status := request.URL.Query().Get("status")
		if status == "" {
			log.Error("no new status provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no new status provided"))
			return
		}

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		userId, err := statusEditor.GetUserId(username)
		if err != nil {
			log.Error("user not found", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("user not found"))
			return
		}

		perms, err := statusEditor.IsHaveBidPerms(userId, bidId)
		if err != nil {
			log.Error("bid not found", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("bid not found"))
			return
		}
		if !perms {
			log.Error("no permission")
			writer.WriteHeader(http.StatusForbidden)
			render.JSON(writer, request, response.Error("no permission"))
			return
		}

		resp, err := statusEditor.EditBidStatus(bidId, status)
		if err != nil {
			log.Error("failed to edit bid status", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			render.JSON(writer, request, response.Error("fail to edit bid status"))
			return
		}

		log.Info("success to edit bid status", slog.Any("response", resp))
		render.JSON(writer, request, resp)
	}
}
