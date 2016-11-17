package models

//go:generate easyjson -no_std_marshalers $GOFILE

import (
	"time"

	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
)

//Item represents one donatable item in a given game
//easyjson:json
type Item struct {
	Key      string                 `json:"item" bson:"key,omitempty"`
	Metadata map[string]interface{} `json:"metadata" bson:"metadata"`

	LimitOfCardsInEachDonationRequest int `json:"limitOfCardsInEachDonationRequest" bson:"limitOfCardsInEachDonationRequest"`
	LimitOfCardsPerPlayerDonation     int `json:"limitOfCardsPerPlayerDonation" bson:"limitOfCardsPerPlayerDonation"`

	//This weight counts for the donation cooldown limits of each player
	WeightPerDonation int `json:"weightPerDonation" bson:"weightPerDonation"`

	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

//NewItem returns a configured new item
func NewItem(
	key string, metadata map[string]interface{},
	weightPerDonation,
	limitOfCardsPerPlayerDonation,
	limitOfCardsInEachDonationRequest int,
) *Item {
	return &Item{
		Key:                               key,
		Metadata:                          metadata,
		WeightPerDonation:                 weightPerDonation,
		LimitOfCardsPerPlayerDonation:     limitOfCardsPerPlayerDonation,
		LimitOfCardsInEachDonationRequest: limitOfCardsInEachDonationRequest,
		UpdatedAt:                         time.Now().UTC(),
	}
}

//ToJSON of this struct
func (i *Item) ToJSON() ([]byte, error) {
	w := jwriter.Writer{}
	i.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

//GetItemFromJSON unmarshals the donation request from the specified JSON
func GetItemFromJSON(data []byte) (*Item, error) {
	donationRequest := &Item{}
	l := jlexer.Lexer{Data: data}
	donationRequest.UnmarshalEasyJSON(&l)
	return donationRequest, l.Error()
}
