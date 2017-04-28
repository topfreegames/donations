package models

//go:generate easyjson -no_std_marshalers $GOFILE

import (
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/donations/errors"
	"github.com/topfreegames/donations/log"
	"github.com/uber-go/zap"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//ResetType represents the type of time that resets scores
type ResetType int

const (
	//NoReset means the value does not reset
	NoReset ResetType = iota
	//DailyReset means the value resets daily
	DailyReset = iota
	//WeeklyReset means the value resets weekly
	WeeklyReset = iota
	//MonthlyReset means the value resets monthly
	MonthlyReset = iota
)

//GetDonationWeightKey returns the key to be used in redis for the scores
func GetDonationWeightKey(prefix, gameID, clanID string, date time.Time, resetType ResetType) string {
	switch resetType {
	case DailyReset:
		return fmt.Sprintf(
			"donations::donation-weight::%s::%s::%s::%s",
			prefix,
			gameID,
			clanID,
			date.Format("2006-01-02"),
		)
	case WeeklyReset:
		isoYear, isoWeek := date.ISOWeek()
		return fmt.Sprintf(
			"donations::donation-weight::%s::%s::%s::%s",
			prefix,
			gameID,
			clanID,
			fmt.Sprintf("%d-%d", isoYear, isoWeek),
		)
	case MonthlyReset:
		return fmt.Sprintf(
			"donations::donation-weight::%s::%s::%s::%s",
			prefix,
			gameID,
			clanID,
			date.Format("2006-01"),
		)
	default:
		return fmt.Sprintf("donations::donation-weight::%s::%s::%s", prefix, gameID, clanID)
	}
}

//GetExpirationDate returns the date to set the expiration in seconds to for the given reset type
func GetExpirationDate(date time.Time, resetType ResetType, clock Clock) int64 {
	switch resetType {
	case DailyReset:
		dt := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		return dt.AddDate(0, 0, 2).Unix() - clock.GetUTCTime().Unix()
	case WeeklyReset:
		day := int(date.Weekday())
		firstDay := date.AddDate(0, 0, (-1*day + 1))
		return firstDay.AddDate(0, 0, 14).Unix() - clock.GetUTCTime().Unix()
	case MonthlyReset:
		dt := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		return dt.AddDate(0, 2, 0).Unix() - clock.GetUTCTime().Unix()
	default:
		return 0
	}
}

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
	ID                string `json:"id" bson:"_id,omitempty"`
	GameID            string `json:"gameID" bson:"gameID"`
	Clan              string `json:"clan" bson:"clan"`
	Player            string `json:"player" bson:"player"`
	DonationRequestID string `json:"donationRequestID" bson:"donationRequestID"`
	Amount            int    `json:"amount" bson:"amount"`
	Weight            int    `json:"weight" bson:"weight"`
	CreatedAt         int64  `json:"createdAt" bson:"createdAt"`
	Metadata          map[string]interface{} `json:"metadata" bson:"metadata"`
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

func (d *DonationRequest) validateDonationCooldownPerPlayer(
	game *Game, player *Player, maxWeightPerPlayer int,
	db *mgo.Database, logger zap.Logger,
) error {
	cooldown := int((time.Duration(game.DonationCooldownHours) * time.Hour).Seconds())
	windowElapsed := int(d.Clock.GetUTCTime().Unix() - player.DonationWindowStart)
	if player.DonationWindowStart == 0 || windowElapsed > cooldown {
		return nil
	}

	from := player.DonationWindowStart
	to := d.Clock.GetUTCTime().Unix()

	totalWeight, err := GetDonationWeightForPlayer(player.ID, from, to, db, logger)
	if err != nil {
		return err
	}
	if totalWeight >= maxWeightPerPlayer {
		return &errors.DonationCooldownViolatedError{
			GameID:               game.ID,
			PlayerID:             player.ID,
			TotalWeightForPeriod: totalWeight,
			MaxWeightForPerior:   maxWeightPerPlayer,
		}
	}
	return nil
}

//Donate an item in a given Donation Request
func (d *DonationRequest) Donate(playerID string, amount, maxWeightPerPlayer int, r redis.Conn, db *mgo.Database, logger zap.Logger) error {
	l := logger.With(
		zap.String("source", "DonationRequestModel"),
		zap.String("operation", "Donate"),
		zap.String("player", playerID),
		zap.Int("amount", amount),
	)

	err := d.validateDonationInputs(playerID, amount, logger)
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

	player, err := GetPlayerByID(playerID, db, logger)
	if err != nil {
		log.E(l, "Could not find player.", func(cm log.CM) {
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

	err = d.ValidateDonationRequestLimitPerPlayer(game, playerID, amount, logger)
	if err != nil {
		log.E(l, err.Error(), func(cm log.CM) {
			cm.Write(zap.Error(err))
		})

		return err
	}

	err = d.validateDonationCooldownPerPlayer(game, player, maxWeightPerPlayer, db, logger)
	if err != nil {
		log.E(l, err.Error(), func(cm log.CM) {
			cm.Write(zap.Error(err))
		})

		return err
	}

	log.D(l, "Saving donation...")
	err = d.insertDonationAndUpdatePlayer(player, amount, game, r, db, l)
	if err != nil {
		log.E(l, err.Error(), func(cm log.CM) {
			cm.Write(zap.Error(err))
		})

		return err
	}

	log.D(l, "Donation added successfully.")

	return nil
}

func (d *DonationRequest) insertDonationAndUpdatePlayer(
	player *Player, amount int,
	game *Game,
	r redis.Conn,
	db *mgo.Database, l zap.Logger,
) error {
	log.D(l, "Saving donation...")

	txID := uuid.NewV4().String()
	rb := func() error {
		err := GetDonationRequestsCollection(db).Update(
			bson.M{"_id": d.ID},
			bson.M{"$pull": bson.M{"donations": bson.M{"$eq": []string{"txID", txID}}}},
		)
		return err
	}

	item := game.Items[d.Item]

	donation := Donation{
		ID:                uuid.NewV4().String(),
		GameID:            d.GameID,
		Clan:              d.Clan,
		Player:            player.ID,
		DonationRequestID: d.ID,
		Amount:            amount,
		Weight:            item.WeightPerDonation,
		CreatedAt:         d.Clock.GetUTCTime().Unix(),
		Metadata: 		   d.Metadata
	}
	d.Donations = append(d.Donations, donation)

	set := bson.M{
		"updatedAt": d.Clock.GetUTCTime().Unix(),
	}

	if d.GetDonationCount()+amount >= item.LimitOfItemsInEachDonationRequest {
		set["finishedAt"] = d.Clock.GetUTCTime().Unix()
	}

	query := bson.M{"_id": d.ID}
	update := bson.M{
		"$set": set,
		"$push": bson.M{"donations": bson.M{
			"_id":       donation.ID,
			"gameID":    donation.GameID,
			"clan":      donation.Clan,
			"player":    donation.Player,
			"amount":    donation.Amount,
			"weight":    item.WeightPerDonation,
			"createdAt": donation.CreatedAt,
			"metadata":  donation.Metadata,
			"txID":      txID,
		}},
	}

	err := GetDonationRequestsCollection(db).Update(query, update)
	if err != nil {
		d.Donations = d.Donations[:len(d.Donations)-1]
		log.E(l, "Failed to add donation.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	err = GetDonationsCollection(db).Insert(donation)
	if err != nil {
		log.E(l, "Failed to add donation.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		err = rb()
		return err
	}

	//Number of seconds in playerDonationWindowHours
	cooldown := int((time.Duration(game.DonationCooldownHours) * time.Hour).Seconds())

	//If player.DonationWindowStart is zero, player has never donated before
	noWindow := player.DonationWindowStart == 0

	//Amount of seconds elapsed since player donation window started
	//If player has no donation window, this will be ignored
	windowElapsed := int(d.Clock.GetUTCTime().Unix() - player.DonationWindowStart)

	//If no window found for player or last window is older than player window cooldown, update it
	if noWindow || windowElapsed > cooldown {
		err = UpdateDonationWindowStart(game.ID, player.ID, d.Clock.GetUTCTime().Unix(), db, l)
		if err != nil {
			d.Donations = d.Donations[:len(d.Donations)-1]
			log.E(l, "Failed to set player update window start.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			err = rb()
			return err
		}
	}

	if d.Clan != "" {
		err = IncrementDonationWeightForClan(r, d.GameID, d.Clan, donation.Weight, d.Clock)
		if err != nil {
			err = rb()
			return err
		}
	}

	err = IncrementDonationWeightForPlayer(r, d.GameID, donation.Player, donation.Weight, d.Clock)
	if err != nil {
		err = rb()
		return err
	}

	return nil
}

//ValidateDonationRequestLimit ensures that no more than the allowed number of donations has been donated
func (d *DonationRequest) ValidateDonationRequestLimit(game *Game, amount int, logger zap.Logger) error {
	item := game.Items[d.Item]
	if d.GetDonationCount()+amount > item.LimitOfItemsInEachDonationRequest {
		err := &errors.LimitOfItemsInDonationRequestReachedError{
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

//ValidateDonationRequestLimitPerPlayer ensures that no more than the allowed number of donations has been donated per player
func (d *DonationRequest) ValidateDonationRequestLimitPerPlayer(game *Game, player string, amount int, logger zap.Logger) error {
	item := game.Items[d.Item]
	currentDonations := d.GetDonationCountForPlayer(player)
	if currentDonations+amount > item.LimitOfItemsPerPlayerDonation {
		err := &errors.LimitOfItemsPerPlayerInDonationRequestReachedError{
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

//GetDonationsCollection to update or query games
func GetDonationsCollection(db *mgo.Database) *mgo.Collection {
	return db.C("donations")
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

//GetDonationWeightForClan returns the donation weight for a clan in a given interval
func GetDonationWeightForClan(gameID, clanID string, dt time.Time, resetType ResetType, r redis.Conn, logger zap.Logger) (int, error) {
	key := GetDonationWeightKey("clan", gameID, clanID, dt, resetType)
	result, err := r.Do("GET", key)
	if err != nil {
		return 0, err
	}
	if result == nil {
		return 0, nil
	}
	res, err := strconv.ParseInt(string(result.([]uint8)), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(res), nil
}

func GetDonationRequestsCollectionForClan(gameID, clanID string, db *mgo.Databaser redis.Conn, logger zap.Logger) ([]*DonationRequest, error)
{
	var requests []*DonationRequest
	err := GetDonationRequestsCollection(db).Find(bson.M{"clan" : clanID, "gameID" : gameID}).All(&requests)

	if(err != nil){
		return nil, err
	}

	return requests, nil
}

//IncrementDonationWeightForClan should increment the donation weight for a clan for all time periods
func IncrementDonationWeightForClan(redis redis.Conn, gameID, clanID string, weight int, clock Clock) error {
	return incrementDonationWeight(redis, "clan", gameID, clanID, weight, clock)
}

//IncrementDonationWeightForPlayer should increment the donation weight for a player for all time periods
func IncrementDonationWeightForPlayer(redis redis.Conn, gameID, playerID string, weight int, clock Clock) error {
	return incrementDonationWeight(redis, "player", gameID, playerID, weight, clock)
}

func incrementDonationWeight(redis redis.Conn, prefix, gameID, id string, weight int, clock Clock) error {
	if redis == nil {
		return fmt.Errorf("The redis client must not be nil and must be connected to redis.")
	}
	dt := clock.GetUTCTime()
	key := GetDonationWeightKey(prefix, gameID, id, dt, NoReset)

	dailyKey := GetDonationWeightKey(prefix, gameID, id, dt, DailyReset)
	dailyExpiration := GetExpirationDate(dt, DailyReset, clock)

	weeklyKey := GetDonationWeightKey(prefix, gameID, id, dt, WeeklyReset)
	weeklyExpiration := GetExpirationDate(dt, WeeklyReset, clock)

	monthlyKey := GetDonationWeightKey(prefix, gameID, id, dt, MonthlyReset)
	monthlyExpiration := GetExpirationDate(dt, MonthlyReset, clock)

	redis.Send("MULTI")

	redis.Send("INCRBY", key, weight)
	redis.Send("INCRBY", dailyKey, weight)
	redis.Send("INCRBY", weeklyKey, weight)
	redis.Send("INCRBY", monthlyKey, weight)
	redis.Send("EXPIRE", dailyKey, dailyExpiration)
	redis.Send("EXPIRE", weeklyKey, weeklyExpiration)
	redis.Send("EXPIRE", monthlyKey, monthlyExpiration)

	_, err := redis.Do("EXEC")
	if err != nil {
		return err
	}
	return nil
}
