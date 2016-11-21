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

//Clock identifies a clock to be used by the model
type Clock interface {
	GetUTCTime() time.Time
}

//RealClock returns the current time as UTC Date
type RealClock struct{}

//GetUTCTime returns the current time as UTC Date
func (r *RealClock) GetUTCTime() time.Time {
	return time.Now().UTC()
}

//Donation represents a specific donation a player made to a request
//easyjson:json
type Donation struct {
	ID        string `json:"id" bson:"_id,omitempty"`
	Player    string `json:"player" bson:"player"`
	Amount    int    `json:"amount" bson:"amount"`
	Weight    int    `json:"weight" bson:"weight"`
	CreatedAt int64  `json:"createdAt" bson:"createdAt"`
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

	Clock Clock `json:"-" bson:"-"`
}

//NewDonationRequest returns a new instance of DonationRequest
func NewDonationRequest(gameID string, item, player, clan string, clock ...Clock) *DonationRequest {
	var cl Clock
	cl = &RealClock{}
	if len(clock) == 1 {
		cl = clock[0]
	}

	return &DonationRequest{
		GameID: gameID,
		Item:   item,
		Player: player,
		Clan:   clan,
		Clock:  cl,
	}
}

//GetGame associated to this donation request
func (d *DonationRequest) GetGame(db *mgo.Database, logger zap.Logger) (*Game, error) {
	return GetGameByID(d.GameID, db, logger)
}

func (d *DonationRequest) validateDonationRequest(logger zap.Logger) error {
	l := logger.With(
		zap.String("source", "DonationRequestModel"),
		zap.String("operation", "validateDonationRequest"),
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

	return nil
}

func (d *DonationRequest) validateDonationRequestCooldown(gameID string, cooldown int, db *mgo.Database, logger zap.Logger) error {
	currentTime := d.Clock.GetUTCTime()
	c, err := GetDonationRequestsCollection(db).Find(bson.M{
		"player": d.Player,
		"createdAt": bson.M{
			"$gte": (currentTime.Add(-1 * (time.Duration(cooldown) * time.Hour))).Unix(),
		},
	}).Count()
	if err != nil {
		return err
	}

	if c > 0 {
		return &errors.DonationRequestCooldownViolatedError{
			Time:    currentTime.Unix(),
			GameID:  gameID,
			ItemKey: d.Item,
		}
	}

	return nil
}

//Create new donation request
func (d *DonationRequest) Create(db *mgo.Database, logger zap.Logger) error {
	l := logger.With(
		zap.String("source", "DonationRequestModel"),
		zap.String("operation", "Save"),
	)

	err := d.validateDonationRequest(logger)
	if err != nil {
		return err
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

	cooldown := game.DonationRequestCooldownHours
	err = d.validateDonationRequestCooldown(game.ID, cooldown, db, logger)
	if err != nil {
		log.E(l, "Donation cooldown infringed.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	d.ID = uuid.NewV4().String()
	d.CreatedAt = d.Clock.GetUTCTime().Unix()
	d.UpdatedAt = d.Clock.GetUTCTime().Unix()

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

func (d *DonationRequest) validateDonationInputs(player string, amount int, logger zap.Logger) error {
	l := logger.With(
		zap.String("source", "DonationRequestModel"),
		zap.String("operation", "Save"),
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

	err := d.validateDonationInputs(player, amount, logger)
	if err != nil {
		return err
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

	err = d.ValidateDonationRequestLimitPerPlayer(game, player, amount, logger)
	if err != nil {
		log.E(l, err.Error(), func(cm log.CM) {
			cm.Write(zap.Error(err))
		})

		return err
	}

	d.Donations = append(d.Donations, Donation{Player: player, Amount: amount})

	set := bson.M{
		"updatedAt": d.Clock.GetUTCTime().Unix(),
	}

	item := game.Items[d.Item]
	if d.GetDonationCount()+amount >= item.LimitOfCardsInEachDonationRequest {
		set["finishedAt"] = d.Clock.GetUTCTime().Unix()
	}

	log.D(l, "Saving donation...")
	query := bson.M{"_id": d.ID}
	update := bson.M{
		"$set": set,
		"$push": bson.M{"donations": bson.M{
			"_id":       uuid.NewV4().String(),
			"player":    player,
			"amount":    amount,
			"weight":    item.WeightPerDonation,
			"createdAt": d.Clock.GetUTCTime().Unix(),
		}},
	}

	err = GetDonationRequestsCollection(db).Update(query, update)
	if err != nil {
		log.E(l, "Failed to add donation.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	err = UpdateDonationWindowStart(game.ID, player, time.Now().UTC().Unix(), db, logger)
	if err != nil {
		log.E(l, "Failed to set player update window start.", func(cm log.CM) {
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
			ItemKey:           d.Item,
			Amount:            amount,
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

//ValidateDonationRequestLimitPerPlayer ensures that no more than the allowed number of cards has been donated per player
func (d *DonationRequest) ValidateDonationRequestLimitPerPlayer(game *Game, player string, amount int, logger zap.Logger) error {
	item := game.Items[d.Item]
	currentDonations := d.GetDonationCountForPlayer(player)
	if currentDonations+amount > item.LimitOfCardsPerPlayerDonation {
		err := &errors.LimitOfCardsPerPlayerInDonationRequestReachedError{
			GameID:               game.ID,
			DonationRequestID:    d.ID,
			ItemKey:              d.Item,
			Amount:               amount,
			Player:               player,
			CurrentDonationCount: currentDonations,
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

//GetDonationCountForPlayer returns the total amount of donations for a given player ID
func (d *DonationRequest) GetDonationCountForPlayer(playerID string) int {
	sum := 0
	for i := 0; i < len(d.Donations); i++ {
		if d.Donations[i].Player == playerID {
			sum += d.Donations[i].Amount
		}
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
	donationRequest.Clock = &RealClock{}
	return &donationRequest, nil
}

//GetDonationWeightForPlayer returns the donation weight for a given player in a given interval
func GetDonationWeightForPlayer(playerID string, from, to int64, db *mgo.Database, logger zap.Logger) (int, error) {
	coll := GetDonationRequestsCollection(db)

	query := []bson.M{
		bson.M{
			"$match": bson.M{
				"donations.player":    playerID,
				"donations.createdAt": bson.M{"$gte": from, "$lte": to},
			},
		},
		bson.M{"$unwind": "$donations"},
		bson.M{
			"$project": bson.M{
				"_id":       "$donations._id",
				"player":    "$donations.player",
				"weight":    "$donations.weight",
				"amount":    "$donations.amount",
				"createdAt": "$donations.createdAt",
			},
		},
		bson.M{
			"$match": bson.M{
				"createdAt": bson.M{"$gte": from, "$lte": to},
			},
		},
		bson.M{
			"$group": bson.M{
				"_id":         "$player",
				"player":      bson.M{"$first": "$player"},
				"totalWeight": bson.M{"$sum": "$weight"},
			},
		},
	}

	pipe := coll.Pipe(query)
	resp := []bson.M{}
	err := pipe.All(&resp)
	if err != nil {
		return 0, err
	}

	if len(resp) == 0 {
		return 0, nil
	}

	return resp[0]["totalWeight"].(int), nil
}
