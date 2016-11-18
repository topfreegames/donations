package models

//go:generate easyjson -no_std_marshalers $GOFILE

import (
	"fmt"
	"time"

	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
	uuid "github.com/satori/go.uuid"
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
	GameID    string     `json:"gameID" bson:"gameID"`
	Donations []Donation `json:"donations" bson:""`

	CreatedAt  int64 `json:"createdAt" bson:"createdAt"`
	UpdatedAt  int64 `json:"updatedAt" bson:"updatedAt,omitempty"`
	FinishedAt int64 `json:"finishedAt" bson:"finishedAt,omitempty"`
}

//GetGame associated to this donation request
func (d *DonationRequest) GetGame(db *mgo.Database, logger zap.Logger) (*Game, error) {
	return GetGameByID(d.GameID, db, logger)
}

//Create new donation request
func (d *DonationRequest) Create(db *mgo.Database, logger zap.Logger) error {
	l := logger.With(
		zap.String("source", "DonationRequestModel"),
		zap.String("operation", "Save"),
	)

	if d.GameID == "" {
		err := &errors.ParameterIsRequiredError{
			Parameter: "GameID",
			Model:     "DonationRequest",
		}

		log.E(l, "GameID cannot be null", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})

		return err
	}

	if d.Item == "" {
		return &errors.ParameterIsRequiredError{
			Parameter: "Item",
			Model:     "DonationRequest",
		}
	}

	game, err := d.GetGame(db, logger)
	if err != nil {
		log.E(l, "Failed to get game associated to Donation Request.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	if _, ok := game.Items[d.Item]; !ok {
		return &errors.ItemNotFoundInGameError{
			ItemKey: d.Item,
			GameID:  game.ID,
		}
	}

	d.ID = uuid.NewV4().String()
	d.CreatedAt = time.Now().UTC().Unix()
	d.UpdatedAt = time.Now().UTC().Unix()

	log.D(l, "Saving donation request...")
	err = GetDonationRequestsCollection(db).Insert(d)

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
func (d *DonationRequest) Donate(player string, amount int, db *mgo.Database, logger zap.Logger) error {
	l := logger.With(
		zap.String("source", "DonationRequestModel"),
		zap.String("operation", "Donate"),
		zap.String("player", player),
		zap.Int("amount", amount),
	)

	if d.ID == "" {
		err := fmt.Errorf("Can't donate without a proper DB-loaded DonationRequest (ID is required).")
		log.E(l, "DonationRequest must be loaded", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})

		return err
	}

	if player == "" {
		return &errors.ParameterIsRequiredError{
			Parameter: "Player",
			Model:     "Donation",
		}
	}

	if amount <= 0 {
		return &errors.ParameterIsRequiredError{
			Parameter: "Amount",
			Model:     "Donation",
		}
	}

	game, err := d.GetGame(db, logger)
	if err != nil {
		log.E(l, "DonationRequest must be loaded", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})

		return err
	}

	err = d.ValidateDonationRequestLimit(game, amount, logger)
	if err != nil {
		log.E(l, err.Error(), func(cm log.CM) {
			cm.Write(zap.Error(err))
		})

		return err
	}

	d.Donations = append(d.Donations, Donation{Player: player, Amount: amount})

	set := bson.M{
		"updatedAt": time.Now().UTC().Unix(),
	}

	if d.GetDonationCount()+1 >= game.Items[d.Item].LimitOfCardsInEachDonationRequest {
		set["finishedAt"] = time.Now().UTC().Unix()
	}

	log.D(l, "Saving donation...")
	query := bson.M{"_id": d.ID}
	update := bson.M{
		"$set": set,
		"$push": bson.M{"donations": bson.M{
			"player": player,
			"amount": amount,
		}},
	}

	err = GetDonationRequestsCollection(db).Update(query, update)
	if err != nil {
		log.E(l, "Failed to add donation.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	log.D(l, "Donation added successfully.")

	return nil
}

//ValidateDonationRequestLimit ensures that no more than the allowed number of cards has been donated
func (d *DonationRequest) ValidateDonationRequestLimit(game *Game, amount int, logger zap.Logger) error {
	item := game.Items[d.Item]
	if d.GetDonationCount()+amount > item.LimitOfCardsInEachDonationRequest {
		err := &errors.LimitOfCardsInDonationRequestReachedError{
			GameID:            game.ID,
			DonationRequestID: d.ID,
		}
		log.E(logger, err.Error(), func(cm log.CM) {
			cm.Write(zap.Error(err))
			cm.Write(zap.String("GameID", game.ID))
			cm.Write(zap.String("DonationRequestID", d.ID))
		})

		return err
	}
	return nil
}

//GetDonationCount returns the total amount of donations
func (d *DonationRequest) GetDonationCount() int {
	sum := 0
	for i := 0; i < len(d.Donations); i++ {
		sum += d.Donations[i].Amount
	}
	return sum
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
func NewDonationRequest(gameID string, item, player, clan string) *DonationRequest {
	return &DonationRequest{
		GameID: gameID,
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
