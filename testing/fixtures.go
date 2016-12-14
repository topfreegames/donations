package testing

import (
	"database/sql"
	"fmt"

	"math/rand"

	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/donations/models"
	"github.com/uber-go/zap"
	mgo "gopkg.in/mgo.v2"
)

//GetTestGame to use in tests
func GetTestGame(db *mgo.Database, logger zap.Logger, withItems bool, options ...map[string]interface{}) (*models.Game, error) {
	opt := map[string]interface{}{
		"LimitOfItemsInEachDonationRequest": 6,
		"LimitOfItemsPerPlayerDonation":     2,
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
		opt["DonationCooldownHours"].(int),        // DonationCooldownHours
		opt["DonationRequestCooldownHours"].(int), // DonationRequestCooldownHours
	)
	err := game.Save(db, logger)
	if err != nil {
		return nil, err
	}

	if withItems {
		for i := 0; i < 10; i++ {
			_, err = game.AddItem(
				fmt.Sprintf("item-%d", i), map[string]interface{}{"x": i},
				opt["LimitOfItemsInEachDonationRequest"].(int),
				opt["LimitOfItemsPerPlayerDonation"].(int),
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

func buildTestDonationRequest(game *models.Game, clanID string) *models.DonationRequest {
	return models.NewDonationRequest(
		game.ID,
		GetFirstItem(game).Key,
		uuid.NewV4().String(),
		clanID,
	)
}

//GetTestDonationRequest to use in tests
func GetTestDonationRequest(game *models.Game, db *mgo.Database, logger zap.Logger) (*models.DonationRequest, error) {
	donationRequest := buildTestDonationRequest(game, uuid.NewV4().String())
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
	err := models.GetPlayersCollection(db).Insert(player)
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

func getRandomPlayers(game *models.Game, numberOfPlayers int, db *mgo.Database) ([]*models.Player, error) {
	players := []*models.Player{}

	for i := 0; i < numberOfPlayers; i++ {
		player := models.NewPlayer(
			game.ID,
			uuid.NewV4().String(),
			0,
		)
		err := models.GetPlayersCollection(db).Insert(player)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}
	return players, nil
}

//GetBulkDonations gets many donations and donation requests
func GetBulkDonations(game *models.Game, clan string, donationRequestCount, donationsPerDonationRequest int, timestamp int64, db *mgo.Database, logger zap.Logger) ([]*models.Player, []*models.DonationRequest, error) {
	players, err := getRandomPlayers(game, 50, db)
	if err != nil {
		return nil, nil, err
	}

	donationRequests := []*models.DonationRequest{}
	donations := []*models.Donation{}
	for i := 0; i < donationRequestCount; i++ {
		clanID := clan
		if rand.Intn(2) == 1 {
			clanID = uuid.NewV4().String()
		}
		donationRequest := buildTestDonationRequest(game, clanID)
		donationRequest.ID = uuid.NewV4().String()
		donationRequest.CreatedAt = timestamp
		donationRequest.UpdatedAt = timestamp

		testPlayer := players[rand.Intn(len(players))]
		donationRequest.Donations = []models.Donation{}
		donationRequest.Player = testPlayer.ID
		donationRequests = append(donationRequests, donationRequest)

		for j := 0; j < donationsPerDonationRequest; j++ {
			donationPlayer := players[rand.Intn(len(players))]
			for donationPlayer.ID == testPlayer.ID {
				donationPlayer = players[rand.Intn(len(players))]
			}
			donation := models.Donation{
				GameID:            donationRequest.GameID,
				Clan:              donationRequest.Clan,
				Player:            donationPlayer.ID,
				DonationRequestID: donationRequest.ID,
				Weight:            1,
				Amount:            1,
				CreatedAt:         timestamp,
			}
			//donationRequest.Donations = append(donationRequest.Donations, donation)
			donations = append(donations, &donation)
		}
	}

	db.Session.Refresh()
	coll := models.GetDonationRequestsCollection(db)
	bulk := coll.Bulk()
	for _, dr := range donationRequests {
		bulk.Insert(dr)
	}
	_, err = bulk.Run()
	if err != nil {
		return nil, nil, err
	}

	db.Session.Refresh()
	coll2 := models.GetDonationsCollection(db)

	index := mgo.Index{
		Key:        []string{"gameID", "clan", "createdAt"},
		Unique:     false,
		DropDups:   false,
		Background: true, // See notes.
		Sparse:     true,
	}
	err = coll2.EnsureIndex(index)
	if err != nil {
		return nil, nil, err
	}

	bulk2 := coll2.Bulk()

	for _, d := range donations {
		bulk2.Insert(d)
	}
	_, err = bulk2.Run()
	if err != nil {
		return nil, nil, err
	}

	return players, donationRequests, nil
}
