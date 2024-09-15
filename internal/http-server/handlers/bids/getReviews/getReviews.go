package getReviews

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"zadanie-6105/internal/lib/api/DTO"
	"zadanie-6105/internal/lib/api/response"
)

type ReviewsGetter interface {
	GetReviews(bids []DTO.BidResponse, limit, offset string) ([]DTO.FeedbackResponse, error)
	GetUserId(username string) (string, error)
	IsHaveTenderPerms(userId, tenderId string) (bool, error)
	GetPersonalBids(limit string, offset string, userId string) ([]DTO.BidResponse, error)
}

func New(log *slog.Logger, reviewsGetter ReviewsGetter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.tender.getReviews"

		tenderId := chi.URLParam(request, "tenderId")
		if tenderId == "" {
			log.Error("no tenderId provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no tenderId provided"))
			return
		}
		limit := request.URL.Query().Get("limit")
		if limit == "" {
			limit = "5"
		}

		offset := request.URL.Query().Get("offset")
		if offset == "" {
			offset = "0"
		}

		authorUsername := request.URL.Query().Get("authorUsername")
		if authorUsername == "" {
			log.Error("no author username provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no author username provided"))
			return
		}

		requesterUsername := request.URL.Query().Get("requesterUsername")
		if requesterUsername == "" {
			log.Error("no requester username  provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no requester username  provided"))
			return
		}

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		requesterId, err := reviewsGetter.GetUserId(requesterUsername)
		if err != nil {
			log.Error("no requester user found")
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("no requester user with this nickname"))
			return
		}

		authorId, err := reviewsGetter.GetUserId(authorUsername)
		if err != nil {
			log.Error("no author found")
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("no author with this nickname"))
			return
		}

		perms, err := reviewsGetter.IsHaveTenderPerms(requesterId, tenderId)
		if err != nil || !perms {
			log.Error("no permission")
			writer.WriteHeader(http.StatusForbidden)
			render.JSON(writer, request, response.Error("no permission"))
			return
		}

		bids, err := reviewsGetter.GetPersonalBids(limit, "0", authorId)

		if len(bids) == 0 {
			render.JSON(writer, request, "user has not received any reviews yet")
			return
		}

		resp, err := reviewsGetter.GetReviews(bids, limit, offset)
		if err != nil {
			log.Error("failed to get user reviews", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			render.JSON(writer, request, response.Error("failed to get user reviews"))
			return
		}

		log.Info("success to get user reviews", slog.Any("response", resp))
		render.JSON(writer, request, resp)
	}
}
