package models

//go:generate easyjson -no_std_marshalers $GOFILE

import (
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
	ID      string                 `json:"id" bson:"_id,omitempty"`
	Name    string                 `json:"name" bson:"name,omitempty"`
	Options map[string]interface{} `json:"options" bson:"options,omitempty"`
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
			"_id":     g.ID,
			"name":    g.Name,
			"options": g.Options,
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

//ToJSON returns the game as JSON
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
func NewGame(name, id string, options map[string]interface{}) *Game {
	return &Game{
		Name:    name,
		ID:      id,
		Options: options,
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
