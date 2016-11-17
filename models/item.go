package models

//go:generate easyjson -no_std_marshalers $GOFILE

import (
	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
)

//Item represents one donatable item in a given game
//easyjson:json
type Item struct {
	Key      string                 `json:"item" bson:"key,omitempty"`
	Metadata map[string]interface{} `json:"metadata" bson:"metadata"`
}

//NewItem returns a configured new item
func NewItem(key string, metadata map[string]interface{}) *Item {
	return &Item{
		Key:      key,
		Metadata: metadata,
	}
}

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
