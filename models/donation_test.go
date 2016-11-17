package models_test

import (
	"fmt"
	"time"

	mgo "gopkg.in/mgo.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/donations/models"
	. "github.com/topfreegames/donations/testing"
	"github.com/uber-go/zap"
)

var _ = Describe("Donation Model", func() {
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

	Describe("Creating a donation request", func() {
		Describe("Feature", func() {
			It("Should create a new donation request", func() {
				start := time.Now().UTC()
				game, err := GetTestGame(db, logger, true)
				Expect(err).NotTo(HaveOccurred())

				itemKey := GetFirstItem(game).Key
				playerID := uuid.NewV4().String()
				clanID := uuid.NewV4().String()
				donationRequest := models.NewDonationRequest(
					game,
					itemKey,
					playerID,
					clanID,
				)
				err = donationRequest.Create(db, logger)
				Expect(err).NotTo(HaveOccurred())

				var dbDonationRequest *models.DonationRequest
				c := models.GetDonationRequestsCollection(db)
				err = c.FindId(donationRequest.ID).One(&dbDonationRequest)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbDonationRequest.Item).To(Equal(itemKey))
				Expect(dbDonationRequest.Player).To(Equal(playerID))
				Expect(dbDonationRequest.Clan).To(Equal(clanID))
				Expect(dbDonationRequest.Game).NotTo(BeNil())
				Expect(dbDonationRequest.Game.ID).To(Equal(game.ID))
				Expect(dbDonationRequest.CreatedAt.Unix()).To(BeNumerically(">", start.Unix()-10))
				Expect(dbDonationRequest.UpdatedAt.Unix()).To(BeNumerically(">", start.Unix()-10))
			})

			It("Should fail to create a new donation request without a game", func() {
				itemID := uuid.NewV4().String()
				playerID := uuid.NewV4().String()
				clanID := uuid.NewV4().String()
				donationRequest := models.NewDonationRequest(
					nil,
					itemID,
					playerID,
					clanID,
				)
				err := donationRequest.Create(db, logger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Game is required to create a new DonationRequest"))
			})

			It("Should fail to create a new donation request with invalid item", func() {
				game, err := GetTestGame(db, logger, true)
				Expect(err).NotTo(HaveOccurred())

				itemID := uuid.NewV4().String()
				playerID := uuid.NewV4().String()
				clanID := uuid.NewV4().String()
				donationRequest := models.NewDonationRequest(
					game,
					itemID,
					playerID,
					clanID,
				)
				err = donationRequest.Create(db, logger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(
					fmt.Sprintf("Item %s was not found in game %s", itemID, game.ID),
				))
			})
		})

		Describe("Measure", func() {
			var game *models.Game
			var err error
			BeforeOnce(func() {
				game, err = GetTestGame(db, logger, true)
				Expect(err).NotTo(HaveOccurred())
			})

			Measure("it should create donation requests fast", func(b Benchmarker) {
				runtime := b.Time("runtime", func() {
					itemKey := GetFirstItem(game).Key
					playerID := uuid.NewV4().String()
					clanID := uuid.NewV4().String()
					donationRequest := models.NewDonationRequest(
						game,
						itemKey,
						playerID,
						clanID,
					)
					err := donationRequest.Create(db, logger)
					Expect(err).NotTo(HaveOccurred())
				})

				Expect(runtime.Seconds()).Should(BeNumerically("<", 0.5), "Operation shouldn't take this long.")
			}, 500)
		})
	})

	Describe("Donating an item", func() {
		Describe("Feature", func() {
			It("Should append a new donation", func() {
				game, err := GetTestGame(db, logger, true)
				Expect(err).NotTo(HaveOccurred())

				itemID := GetFirstItem(game).Key
				playerID := uuid.NewV4().String()
				clanID := uuid.NewV4().String()
				donationRequest := models.NewDonationRequest(
					game,
					itemID,
					playerID,
					clanID,
				)
				err = donationRequest.Create(db, logger)
				Expect(err).NotTo(HaveOccurred())

				donerID := uuid.NewV4().String()
				err = donationRequest.Donate(db, donerID, 1, logger)
				Expect(err).NotTo(HaveOccurred())

				var dbDonationRequest *models.DonationRequest
				c := models.GetDonationRequestsCollection(db)
				err = c.FindId(donationRequest.ID).One(&dbDonationRequest)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbDonationRequest.Donations).To(HaveLen(1))

				Expect(dbDonationRequest.Donations[0].Player).To(Equal(donerID))
				Expect(dbDonationRequest.Donations[0].Amount).To(Equal(1))
			})
		})

		Describe("Measure", func() {
			var game *models.Game
			var donationRequest *models.DonationRequest
			var err error

			BeforeOnce(func() {
				game, err = GetTestGame(db, logger, true)
				Expect(err).NotTo(HaveOccurred())

				itemID := GetFirstItem(game).Key
				playerID := uuid.NewV4().String()
				clanID := uuid.NewV4().String()
				donationRequest = models.NewDonationRequest(
					game,
					itemID,
					playerID,
					clanID,
				)
				err = donationRequest.Create(db, logger)
				Expect(err).NotTo(HaveOccurred())
			})

			Measure("it should donate items fast", func(b Benchmarker) {
				runtime := b.Time("runtime", func() {
					donorID := uuid.NewV4().String()
					err := donationRequest.Donate(
						db,
						donorID,
						1,
						logger,
					)
					Expect(err).NotTo(HaveOccurred())
				})

				Expect(runtime.Seconds()).Should(BeNumerically("<", 0.5), "Operation shouldn't take this long.")
			}, 500)
		})
	})

	Describe("Getting a donation request", func() {
		Describe("Feature", func() {
			It("Should get a donation request by id", func() {
				game, err := GetTestGame(db, logger, true)
				Expect(err).NotTo(HaveOccurred())

				donationRequest, err := GetTestDonationRequest(game, db, logger)
				Expect(err).NotTo(HaveOccurred())

				dbDonationRequest, err := models.GetDonationRequestByID(donationRequest.ID, db, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbDonationRequest.Item).To(Equal(donationRequest.Item))
				Expect(dbDonationRequest.Player).To(Equal(donationRequest.Player))
				Expect(dbDonationRequest.Clan).To(Equal(donationRequest.Clan))
				Expect(dbDonationRequest.Game).NotTo(BeNil())
				Expect(dbDonationRequest.Game.ID).To(Equal(game.ID))
			})
		})

		Describe("Measure", func() {
			var donationRequest *models.DonationRequest

			BeforeOnce(func() {
				game, err := GetTestGame(db, logger, true)
				Expect(err).NotTo(HaveOccurred())

				donationRequest, err = GetTestDonationRequest(game, db, logger)
				Expect(err).NotTo(HaveOccurred())
			})

			Measure("it should get donation requests fast", func(b Benchmarker) {
				runtime := b.Time("runtime", func() {
					_, err := models.GetDonationRequestByID(donationRequest.ID, db, logger)
					Expect(err).NotTo(HaveOccurred())
				})

				Expect(runtime.Seconds()).Should(BeNumerically("<", 0.5), "Operation shouldn't take this long.")
			}, 500)
		})
	})

})
