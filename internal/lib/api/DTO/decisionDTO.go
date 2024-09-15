package DTO

type DecisionResponse struct {
	Id       string `json:"id"`
	BidId    string `json:"bidId"`
	Decision string `json:"decision"`
	UserId   string `json:"userId"`
	TenderId string `json:"tenderId"`
}
