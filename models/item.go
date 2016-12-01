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

	LimitOfItemsInEachDonationRequest int `json:"limitOfItemsInEachDonationRequest" bson:"limitOfItemsInEachDonationRequest"`
	LimitOfItemsPerPlayerDonation     int `json:"limitOfItemsPerPlayerDonation" bson:"limitOfItemsPerPlayerDonation"`

	//This weight counts for the donation cooldown limits of each player
	WeightPerDonation int `json:"weightPerDonation" bson:"weightPerDonation"`

	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt"`
}

//NewItem returns a configured new item
func NewItem(
	key string, metadata map[string]interface{},
	weightPerDonation,
	limitOfItemsPerPlayerDonation,
	limitOfItemsInEachDonationRequest int,
) *Item {
	return &Item{
		Key:                               key,
		Metadata:                          metadata,
		WeightPerDonation:                 weightPerDonation,
		LimitOfItemsPerPlayerDonation:     limitOfItemsPerPlayerDonation,
		LimitOfItemsInEachDonationRequest: limitOfItemsInEachDonationRequest,
		UpdatedAt:                         time.Now().UTC().Unix(),
	}
}

//ToJSON of this struct
func (i *Item) ToJSON() ([]byte, error) {
	w := jwriter.Writer{}
	i.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

//GetItemFromJSON unmarshals the item from the specified JSON
func GetItemFromJSON(data []byte) (*Item, error) {
	obj := &Item{}
	l := jlexer.Lexer{Data: data}
	obj.UnmarshalEasyJSON(&l)
	return obj, l.Error()
}
