package testing

import (
	"database/sql"
	"fmt"

	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/donations/models"
	"github.com/uber-go/zap"
	mgo "gopkg.in/mgo.v2"
)

//GetTestGame to use in tests
func GetTestGame(db *mgo.Database, logger zap.Logger, withItems bool, options ...map[string]interface{}) (*models.Game, error) {
	opt := map[string]interface{}{
		"LimitOfCardsInEachDonationRequest": 6,
		"LimitOfCardsPerPlayerDonation":     2,
		"WeightPerDonation":                 1,
		"DonationRequestCooldownHours":      24,
		"DonationCooldownHours":             8,
	}

	if len(options) == 1 {
		for key, val := range options[0] {
			opt[key] = val
		}
	}

	game := models.NewGame(
		uuid.NewV4().String(),                     // Name
		uuid.NewV4().String(),                     // ID
		opt["DonationRequestCooldownHours"].(int), // DonationRequestCooldownHours
		opt["DonationCooldownHours"].(int),        // DonationCooldownHours
	)
	err := game.Save(db, logger)
	if err != nil {
		return nil, err
	}

	if withItems {
		for i := 0; i < 10; i++ {
			_, err = game.AddItem(
				fmt.Sprintf("item-%d", i), map[string]interface{}{"x": i},
				opt["LimitOfCardsInEachDonationRequest"].(int),
				opt["LimitOfCardsPerPlayerDonation"].(int),
				opt["WeightPerDonation"].(int),
				db, logger,
			)
			if err != nil {
				return nil, err
			}
		}
	}
	return game, nil
}

//GetTestDonationRequest to use in tests
func GetTestDonationRequest(game *models.Game, db *mgo.Database, logger zap.Logger) (*models.DonationRequest, error) {
	donationRequest := models.NewDonationRequest(
		game.ID,
		GetFirstItem(game).Key,
		uuid.NewV4().String(),
		uuid.NewV4().String(),
	)
	err := donationRequest.Create(db, logger)
	return donationRequest, err
}

//GetTestPlayer to use in tests
func GetTestPlayer(game *models.Game, db *mgo.Database, logger zap.Logger) (*models.Player, error) {
	player := models.NewPlayer(
		game.ID,
		uuid.NewV4().String(),
		0,
	)
	err := models.GetDonationRequestsCollection(db).Insert(player)
	return player, err
}

//GetFirstItem In Game
func GetFirstItem(g *models.Game) *models.Item {
	for _, v := range g.Items {
		return &v
	}

	return nil
}

//ToNullInt64 returns valid if int > 0
func ToNullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v > 0}
}
