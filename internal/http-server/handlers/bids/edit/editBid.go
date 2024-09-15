package edit

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	bidOpt "zadanie-6105/internal/lib/api/DTO"
	"zadanie-6105/internal/lib/api/response"
)

type BidEditor interface {
	EditBid(req bidOpt.BidPatchRequest, tenderId string) (bidOpt.BidResponse, error)
	IsHaveBidPerms(userId, bidId string) (bool, error)
	GetUserId(username string) (string, error)
	IsTenderPublic(tenderId string) (bool, error)
}

func New(log *slog.Logger, bidEditor BidEditor) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.bid.editBid"

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

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		var req bidOpt.BidPatchRequest

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

		userId, err := bidEditor.GetUserId(username)
		if err != nil {
			log.Error("user not found", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("user not found"))
			return
		}

		perms, err := bidEditor.IsHaveBidPerms(userId, bidId)
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

		resp, err := bidEditor.EditBid(req, bidId)

		if err != nil {
			log.Error("failed to edit bid", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			render.JSON(writer, request, response.Error("fail to edit bid"))
			return
		}

		log.Info("success to edit bid", slog.Any("response", resp))
		render.JSON(writer, request, resp)
	}
}
