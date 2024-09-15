package DTO

import "time"

type FeedbackResponse struct {
	Id          string    `json:"id"`
	BidFeedback string    `json:"bidfeedback,max=1000"`
	CreatedAt   time.Time `json:"createdAt"`
}
