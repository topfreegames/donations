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
func GetTestGame(db *mgo.Database, logger zap.Logger, withItems bool) (*models.Game, error) {
	game := models.NewGame(
		uuid.NewV4().String(), // Name
		uuid.NewV4().String(), // ID
		24, // DonationRequestCooldownHours
		8,  // DonationCooldownHours
	)
	err := game.Save(db, logger)
	if err != nil {
		return nil, err
	}

	if withItems {
		for i := 0; i < 10; i++ {
			_, err = game.AddItem(fmt.Sprintf("item-%d", i), map[string]interface{}{"x": i}, i, i*2, i*3, db, logger)
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
		game,
		uuid.NewV4().String(),
		uuid.NewV4().String(),
		uuid.NewV4().String(),
	)
	err := donationRequest.Create(db, logger)
	return donationRequest, err
}

//ToNullInt64 returns valid if int > 0
func ToNullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v > 0}
}
