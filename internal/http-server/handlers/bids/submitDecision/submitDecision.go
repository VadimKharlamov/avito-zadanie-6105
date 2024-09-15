package submitDecision

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	bidOpt "zadanie-6105/internal/lib/api/DTO"
	"zadanie-6105/internal/lib/api/response"
)

type DecisionMaker interface {
	SubmitDecision(userId, bidId, decision, tenderId string) (bidOpt.DecisionResponse, error)
	IsHaveBidPerms(userId, bidId string) (bool, error)
	GetUserId(username string) (string, error)
	GetTenderId(bidId string) (string, error)
	IsHaveTenderPerms(userId, tenderId string) (bool, error)
}

func New(log *slog.Logger, decisionMaker DecisionMaker) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.bid.submitDecision"

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

		decision := request.URL.Query().Get("decision")
		if decision == "" && !(decision == "Approved" || decision == "Rejected") {
			log.Error("no correct decision provided")
			writer.WriteHeader(http.StatusBadRequest)
			render.JSON(writer, request, response.Error("no correct decision provided"))
			return
		}

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		userId, err := decisionMaker.GetUserId(username)
		if err != nil {
			log.Error("user not found", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("user not found"))
			return
		}

		tenderId, err := decisionMaker.GetTenderId(bidId)
		if err != nil {
			log.Error("bid not found", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusNotFound)
			render.JSON(writer, request, response.Error("bid not found"))
			return
		}
		userPerms, err := decisionMaker.IsHaveTenderPerms(userId, tenderId)
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

		resp, err := decisionMaker.SubmitDecision(userId, bidId, decision, tenderId)
		if err != nil {
			log.Error("failed to submit decision", slog.String("error", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			render.JSON(writer, request, response.Error("fail to submit decision or already submitted"))
			return
		}

		log.Info("success submit decision", slog.Any("response", resp))
		render.JSON(writer, request, resp)
	}
}
