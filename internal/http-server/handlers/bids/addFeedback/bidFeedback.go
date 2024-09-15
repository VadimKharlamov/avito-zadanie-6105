package addFeedback

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	bidOpt "zadanie-6105/internal/lib/api/DTO"
	"zadanie-6105/internal/lib/api/response"
)

type FeedbackMaker interface {
	CreateFeedback(userId, bidId, bidFeedback string) (bidOpt.FeedbackResponse, error)
	GetTenderId(bidId string) (string, error)
	GetUserId(username string) (string, error)
	IsHaveTenderPerms(userId, tenderId string) (bool, error)
}

func New(log *slog.Logger, feedbackMaker FeedbackMaker) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.bid.feedbackMaker"

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

		bidFeedback := request.URL.Query().Get("bidFeedback")
		if bidFeedback == "" {
			log.Error("no feedback provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no feedback provided"))
			return
		}

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		userId, err := feedbackMaker.GetUserId(username)
		if err != nil {
			log.Error("user not found", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("user not found"))
			return
		}

		tenderId, err := feedbackMaker.GetTenderId(bidId)
		if err != nil {
			log.Error("bid not found", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("bid not found"))
			return
		}
		userPerms, err := feedbackMaker.IsHaveTenderPerms(userId, tenderId)
		if err != nil {
			log.Error("tender not found", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("tender not found"))
			return
		}
		if !userPerms {
			log.Error("no permission")
			writer.WriteHeader(http.StatusForbidden)
			render.JSON(writer, request, response.Error("no permission"))
			return
		}

		resp, err := feedbackMaker.CreateFeedback(userId, bidId, bidFeedback)
		if err != nil {
			log.Error("failed to create feedback", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			render.JSON(writer, request, response.Error("fail to create feedback"))
			return
		}

		log.Info("success create feedback", slog.Any("response", resp))
		render.JSON(writer, request, resp)
	}
}
