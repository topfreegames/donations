package models

//go:generate easyjson -no_std_marshalers $GOFILE

import (
	"fmt"

	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
	"github.com/topfreegames/donations/errors"
	"github.com/topfreegames/donations/log"
	"github.com/uber-go/zap"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Donation represents a specific donation a player made to a request
//easyjson:json
type Donation struct {
	Player string `json:"player" bson:"player"`
	Amount int    `json:"amount" bson:"amount"`
}

//DonationRequest represents a request for an item donation a player made in a game
//easyjson:json
type DonationRequest struct {
	ID        string     `json:"id" bson:"_id,omitempty"`
	Item      string     `json:"item" bson:"item,omitempty"`
	Player    string     `json:"player" bson:"player"`
	Clan      string     `json:"clan" bson:"clan"`
	Game      *Game      `json:"game" bson:"game"`
	Donations []Donation `json:"donations" bson:""`
}

//Create new donation request
func (d *DonationRequest) Create(db *mgo.Database, logger zap.Logger) error {
	l := logger.With(
		zap.String("source", "DonationRequestModel"),
		zap.String("operation", "Save"),
	)

	d.ID = bson.NewObjectId().Hex()

	log.D(l, "Saving donation request...")
	err := GetDonationRequestsCollection(db).Insert(d)

	if err != nil {
		log.E(l, "Failed to save donation request.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	log.D(l, "Donation Request saved successfully.")

	return nil
}

//Donate an item in a given Donation Request
func (d *DonationRequest) Donate(db *mgo.Database, player string, amount int, logger zap.Logger) error {
	l := logger.With(
		zap.String("source", "DonationRequestModel"),
		zap.String("operation", "Donate"),
		zap.String("player", player),
		zap.Int("amount", amount),
	)

	if d.ID == "" {
		return fmt.Errorf("Can't donate without a proper DB-loaded DonationRequest (ID is required).")
	}

	log.D(l, "Saving donation...")
	query := bson.M{"_id": d.ID}
	update := bson.M{"$push": bson.M{"donations": bson.M{
		"player": player,
		"amount": amount,
	}}}

	err := GetDonationRequestsCollection(db).Update(query, update)
	if err != nil {
		log.E(l, "Failed to add donation.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	log.D(l, "Donation added successfully.")

	return nil
}

//ToJSON returns the game as JSON
func (d *DonationRequest) ToJSON() ([]byte, error) {
	w := jwriter.Writer{}
	d.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

//GetDonationRequestFromJSON unmarshals the donation request from the specified JSON
func GetDonationRequestFromJSON(data []byte) (*DonationRequest, error) {
	donationRequest := &DonationRequest{}
	l := jlexer.Lexer{Data: data}
	donationRequest.UnmarshalEasyJSON(&l)
	return donationRequest, l.Error()
}

//GetDonationFromJSON unmarshals the donation from the specified JSON
func GetDonationFromJSON(data []byte) (*Donation, error) {
	donation := &Donation{}
	l := jlexer.Lexer{Data: data}
	donation.UnmarshalEasyJSON(&l)
	return donation, l.Error()
}

//NewDonationRequest returns a new instance of DonationRequest
func NewDonationRequest(game *Game, item, player, clan string) *DonationRequest {
	return &DonationRequest{
		Game:   game,
		Item:   item,
		Player: player,
		Clan:   clan,
	}
}

//GetDonationRequestsCollection to update or query games
func GetDonationRequestsCollection(db *mgo.Database) *mgo.Collection {
	return db.C("requests")
}

//GetDonationRequestByID rtrieves the game by its id
func GetDonationRequestByID(id string, db *mgo.Database, logger zap.Logger) (*DonationRequest, error) {
	var donationRequest DonationRequest
	err := GetDonationRequestsCollection(db).FindId(id).One(&donationRequest)
	if err != nil {
		if err.Error() == NotFoundString {
			return nil, errors.NewDocumentNotFoundError("donationRequest", id)
		}
		return nil, err
	}
	return &donationRequest, nil
}
