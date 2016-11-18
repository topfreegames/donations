package models

//go:generate easyjson -no_std_marshalers $GOFILE

import (
	"time"

	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
	"github.com/topfreegames/donations/errors"
	"github.com/topfreegames/donations/log"
	"github.com/uber-go/zap"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Game represents each game in UR
//easyjson:json
type Game struct {
	ID   string `json:"id" bson:"_id,omitempty"`
	Name string `json:"name" bson:"name,omitempty"`

	// List of items in this game
	Items map[string]Item `json:"items" bson:""`

	// Number of hours since player's first donation before the limit of donations reset.
	// If player donates item A at timestamp X, then their donations limit resets at x + DonationCooldownHours.
	DonationCooldownHours int `json:"donationCooldownHours" bson:"donationCooldownHours"`

	// Cooldown a player must wait before doing his next donation request. Defaults to 8hs
	DonationRequestCooldownHours int `json:"donationRequestCooldownHours" bson:"donationRequestCooldownHours"`

	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

//Save new game if no ID and updates it otherwise
func (g *Game) Save(db *mgo.Database, logger zap.Logger) error {
	l := logger.With(
		zap.String("source", "GameModel"),
		zap.String("operation", "Save"),
	)

	if g.ID == "" {
		g.ID = bson.NewObjectId().String()
	}

	log.D(l, "Saving game...")
	info, err := GetGamesCollection(db).Upsert(
		M{"_id": g.ID},
		M{
			"_id":                          g.ID,
			"name":                         g.Name,
			"items":                        g.Items,
			"donationCooldownHours":        g.DonationCooldownHours,
			"donationRequestCooldownHours": g.DonationRequestCooldownHours,
			"updatedAt":                    time.Now().UTC(),
		},
	)

	if err != nil {
		log.E(l, "Failed to save game.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	log.D(l, "Game saved successfully.", func(cm log.CM) {
		cm.Write(
			zap.Int("UpdatedRows", info.Updated),
			zap.Int("MatchedRows", info.Matched),
			zap.Object("UpsertedID", info.UpsertedId),
		)
	})

	return nil
}

//AddItem to this game
func (g *Game) AddItem(
	key string, metadata map[string]interface{},
	limitOfCardsInEachDonationRequest,
	limitOfCardsPerPlayerDonation,
	weightPerDonation int,
	db *mgo.Database, logger zap.Logger,
) (*Item, error) {
	item := NewItem(
		key, metadata,
		weightPerDonation,
		limitOfCardsPerPlayerDonation,
		limitOfCardsInEachDonationRequest,
	)
	g.Items[key] = *item
	err := g.Save(db, logger)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (g *Game) ToJSON() ([]byte, error) {
	w := jwriter.Writer{}
	g.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

//GetGameFromJSON unmarshals the game from the specified JSON
func GetGameFromJSON(data []byte) (*Game, error) {
	game := &Game{}
	l := jlexer.Lexer{Data: data}
	game.UnmarshalEasyJSON(&l)
	return game, l.Error()
}

//NewGame returns a new instance of Game
func NewGame(
	name, id string,
	donationCooldownHours,
	donationRequestCooldownHours int,
) *Game {
	return &Game{
		Name:  name,
		ID:    id,
		Items: map[string]Item{},
		DonationCooldownHours:        donationCooldownHours,
		DonationRequestCooldownHours: donationRequestCooldownHours,
	}
}

//GetGamesCollection to update or query games
func GetGamesCollection(db *mgo.Database) *mgo.Collection {
	return db.C("games")
}

//GetGameByID rtrieves the game by its id
func GetGameByID(id string, db *mgo.Database, logger zap.Logger) (*Game, error) {
	var game Game
	err := GetGamesCollection(db).FindId(id).One(&game)
	if err != nil {
		if err.Error() == NotFoundString {
			return nil, errors.NewDocumentNotFoundError("games", id)
		}
		return nil, err
	}
	return &game, nil
}
