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

//Player represents one player in a given game
//easyjson:json
type Player struct {
	GameID              string `json:"gameID" bson:"gameID"`
	ID                  string `json:"id" bson:"_id,omitempty"`
	DonationWindowStart int64  `json:"donationWindowStart" bson:"donationWindowStart"`
}

//NewPlayer returns a new player
func NewPlayer(gameID, playerID string, donationWindowStart int64) *Player {
	return &Player{
		GameID:              gameID,
		ID:                  playerID,
		DonationWindowStart: donationWindowStart,
	}
}

//GetPlayersCollection to update or query games
func GetPlayersCollection(db *mgo.Database) *mgo.Collection {
	return db.C("players")
}

//EnsurePlayerExists upserts the player
func EnsurePlayerExists(gameID, id string, db *mgo.Database, logger zap.Logger) error {
	l := logger.With(
		zap.String("source", "PlayerModel"),
		zap.String("operation", "EnsurePlayerExists"),
		zap.String("gameID", gameID),
		zap.String("playerID", id),
	)

	query := bson.M{"_id": id}
	update := bson.M{
		"gameID": gameID,
	}

	_, err := GetPlayersCollection(db).Upsert(query, update)
	if err != nil {
		log.E(l, "Failed to upsert player.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	return nil
}

//UpdateDonationWindowStart updates the donation window start for the player with specified id
func UpdateDonationWindowStart(gameID, id string, timestamp int64, db *mgo.Database, logger zap.Logger) error {
	l := logger.With(
		zap.String("source", "PlayerModel"),
		zap.String("operation", "UpdateDonationWindowStart"),
		zap.String("gameID", gameID),
		zap.String("playerID", id),
		zap.Int64("timestamp", timestamp),
	)

	query := bson.M{"_id": id}
	update := bson.M{
		"donationWindowStart": timestamp,
		"gameID":              gameID,
	}

	_, err := GetPlayersCollection(db).Upsert(query, update)
	if err != nil {
		log.E(l, "Failed to set player donationWindowStart.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	log.D(l, "Player donationWindowStart set successfully.")

	return nil

}

//GetPlayerByID rtrieves the game by its id
func GetPlayerByID(id string, db *mgo.Database, logger zap.Logger) (*Player, error) {
	var player Player
	err := GetPlayersCollection(db).FindId(id).One(&player)
	if err != nil {
		if err.Error() == NotFoundString {
			return nil, errors.NewDocumentNotFoundError("player", id)
		}
		return nil, err
	}
	return &player, nil
}

//ToJSON of this struct
func (p *Player) ToJSON() ([]byte, error) {
	w := jwriter.Writer{}
	p.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

//GetPlayerFromJSON unmarshals the player from the specified JSON
func GetPlayerFromJSON(data []byte) (*Player, error) {
	obj := &Player{}
	l := jlexer.Lexer{Data: data}
	obj.UnmarshalEasyJSON(&l)
	return obj, l.Error()
}
