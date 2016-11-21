package models_test

import (
	"time"

	mgo "gopkg.in/mgo.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/donations/models"
	. "github.com/topfreegames/donations/testing"
	"github.com/uber-go/zap"
)

var _ = Describe("Player Model", func() {
	var logger zap.Logger
	var session *mgo.Session
	var db *mgo.Database

	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)

		session, db = GetTestMongoDB()
	})

	AfterEach(func() {
		session.Close()
		session = nil
		db = nil
	})

	Describe("Player Auxiliary Methods", func() {
		Describe("Serialize/Deserialize", func() {
			Describe("Feature", func() {
				It("Should parse and serialize to json", func() {
					player := &models.Player{
						GameID:              "some-game",
						ID:                  uuid.NewV4().String(),
						DonationWindowStart: time.Now().UTC().Unix(),
					}

					r, err := player.ToJSON()
					Expect(err).NotTo(HaveOccurred())

					rr, err := models.GetPlayerFromJSON(r)
					Expect(err).NotTo(HaveOccurred())

					Expect(rr.ID).To(Equal(player.ID))
					Expect(rr.GameID).To(Equal(player.GameID))
					Expect(rr.DonationWindowStart).To(Equal(player.DonationWindowStart))
				})
			})
		})

		Describe("Update DonationWindowStart", func() {
			Describe("Feature", func() {
				It("Should update when player exists", func() {
					game, err := GetTestGame(db, logger, true)
					Expect(err).NotTo(HaveOccurred())

					player, err := GetTestPlayer(game, db, logger)
					Expect(err).NotTo(HaveOccurred())

					err = models.UpdateDonationWindowStart(game.ID, player.ID, 500, db, logger)
					Expect(err).NotTo(HaveOccurred())

					dbPlayer, err := models.GetPlayerByID(player.ID, db, logger)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.DonationWindowStart).To(Equal(int64(500)))
				})

				It("Should create when player does not exist", func() {
					game, err := GetTestGame(db, logger, true)
					Expect(err).NotTo(HaveOccurred())

					playerID := uuid.NewV4().String()

					err = models.UpdateDonationWindowStart(game.ID, playerID, 500, db, logger)
					Expect(err).NotTo(HaveOccurred())

					dbPlayer, err := models.GetPlayerByID(playerID, db, logger)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.DonationWindowStart).To(Equal(int64(500)))
				})
			})
		})
	})
})
