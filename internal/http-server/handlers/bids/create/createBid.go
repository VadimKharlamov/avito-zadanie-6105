package create

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	bidOpt "zadanie-6105/internal/lib/api/DTO"
	"zadanie-6105/internal/lib/api/response"
)

type BidSaver interface {
	CreateBid(bidData bidOpt.BidRequest) (bidOpt.BidResponse, error)
	IsExist(authorId, authorType string) (bool, error)
	IsTenderPublic(tenderId string) (bool, error)
}

func New(log *slog.Logger, bidSaver BidSaver) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.bid.create"

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		var req bidOpt.BidRequest

		err := render.DecodeJSON(request.Body, &req)
		if err != nil {
			log.Error("failed to deserialize request", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("fail to decode request"))
			return
		}

		log.Info("request bid body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			var validationErrors validator.ValidationErrors
			errors.As(err, &validationErrors)
			log.Error("fail to validate request", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.ValidationError(validationErrors))
			return
		}

		existence, err := bidSaver.IsExist(req.AuthorId, req.AuthorType)
		if err != nil {
			log.Error("failed to check author", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("fail to check author"))
			return
		}
		if !existence {
			log.Error("author not found", slog.String("error", "author not found"))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("author not found"))
			return
		}

		publicStatus, err := bidSaver.IsTenderPublic(req.TenderId)
		if err != nil {
			log.Error("tender not found", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("tender not found"))
			return
		}
		if !publicStatus {
			log.Error("no permission", slog.String("error", "no permission"))
			writer.WriteHeader(http.StatusForbidden)
			render.JSON(writer, request, response.Error("no permission to create bid"))
			return
		}
		resp, err := bidSaver.CreateBid(req)
		if err != nil {
			log.Error("failed to save bid", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusForbidden)
			render.JSON(writer, request, response.Error("fail to save bid"))
			return
		}

		log.Info("success to save bid", slog.Any("response", resp))
		render.JSON(writer, request, resp)
	}
}
