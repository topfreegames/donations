package models_test

import (
	"fmt"
	"time"

	mgo "gopkg.in/mgo.v2"
	redsync "gopkg.in/redsync.v1"

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
	var rs *redsync.Redsync

	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)

		session, db = GetTestMongoDB()
		rs = GetTestRedsync()
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
					game.ID,
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
				Expect(dbDonationRequest.GameID).To(Equal(game.ID))
				Expect(dbDonationRequest.CreatedAt).To(BeNumerically(">", start.Unix()-10))
				Expect(dbDonationRequest.UpdatedAt).To(BeNumerically(">", start.Unix()-10))
				Expect(dbDonationRequest.FinishedAt).To(BeEquivalentTo(0))
			})

			It("Should fail to create a new donation request without a game", func() {
				itemID := uuid.NewV4().String()
				playerID := uuid.NewV4().String()
				clanID := uuid.NewV4().String()
				donationRequest := models.NewDonationRequest(
					"",
					itemID,
					playerID,
					clanID,
				)
				err := donationRequest.Create(db, logger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("GameID is required to create a new DonationRequest"))
			})

			It("Should fail to create a new donation request with null item", func() {
				game, err := GetTestGame(db, logger, true)
				Expect(err).NotTo(HaveOccurred())

				playerID := uuid.NewV4().String()
				clanID := uuid.NewV4().String()
				donationRequest := models.NewDonationRequest(
					game.ID,
					"",
					playerID,
					clanID,
				)
				err = donationRequest.Create(db, logger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Item is required to create a new DonationRequest"))
			})

			It("Should fail to create a new donation request with invalid item", func() {
				game, err := GetTestGame(db, logger, true)
				Expect(err).NotTo(HaveOccurred())

				itemID := uuid.NewV4().String()
				playerID := uuid.NewV4().String()
				clanID := uuid.NewV4().String()
				donationRequest := models.NewDonationRequest(
					game.ID,
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
						game.ID,
						itemKey,
						playerID,
						clanID,
					)
					err := donationRequest.Create(db, logger)
					Expect(err).NotTo(HaveOccurred())
				})

				Expect(runtime.Seconds()).Should(BeNumerically("<", 0.1), "Operation shouldn't take this long.")
			}, 10)
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
					game.ID,
					itemID,
					playerID,
					clanID,
				)
				err = donationRequest.Create(db, logger)
				Expect(err).NotTo(HaveOccurred())

				donerID := uuid.NewV4().String()
				err = donationRequest.Donate(donerID, 1, db, logger)
				Expect(err).NotTo(HaveOccurred())

				dbDonationRequest, err := models.GetDonationRequestByID(donationRequest.ID, db, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbDonationRequest.Donations).To(HaveLen(1))

				Expect(dbDonationRequest.Donations[0].Player).To(Equal(donerID))
				Expect(dbDonationRequest.Donations[0].Amount).To(Equal(1))
			})

			It("Should set the FinishedAt timestamp in the donation once the limit is reached", func() {
				start := time.Now().UTC()
				game, err := GetTestGame(db, logger, true, map[string]interface{}{
					"LimitOfCardsInEachDonationRequest": 2,
				})
				Expect(err).NotTo(HaveOccurred())

				dr, err := GetTestDonationRequest(game, db, logger)

				err = dr.Donate(uuid.NewV4().String(), 2, db, logger)
				Expect(err).NotTo(HaveOccurred())

				dbDonationRequest, err := models.GetDonationRequestByID(dr.ID, db, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbDonationRequest.FinishedAt).To(BeNumerically(">", start.Unix()-10))
			})

			It("Should respect the donation per donation request limit max items", func() {
				game, err := GetTestGame(db, logger, true, map[string]interface{}{
					"LimitOfCardsInEachDonationRequest": 2,
				})
				Expect(err).NotTo(HaveOccurred())

				dr, err := GetTestDonationRequest(game, db, logger)

				err = dr.Donate(uuid.NewV4().String(), 3, db, logger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("This donation request can't accept this donation."))
			})

			It("Should respect the donation per donation request limit", func() {
				game, err := GetTestGame(db, logger, true, map[string]interface{}{
					"LimitOfCardsInEachDonationRequest": 2,
				})
				Expect(err).NotTo(HaveOccurred())

				dr, err := GetTestDonationRequest(game, db, logger)

				err = dr.Donate(uuid.NewV4().String(), 2, db, logger)
				Expect(err).NotTo(HaveOccurred())

				err = dr.Donate(uuid.NewV4().String(), 1, db, logger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("This donation request can't accept this donation."))
			})

			It("Should fail if DonationRequest is not loaded", func() {
				game, err := GetTestGame(db, logger, true, map[string]interface{}{
					"LimitOfCardsInEachDonationRequest": 2,
				})
				Expect(err).NotTo(HaveOccurred())

				itemID := GetFirstItem(game).Key
				dr := models.NewDonationRequest(game.ID, itemID, uuid.NewV4().String(), uuid.NewV4().String())

				err = dr.Donate(uuid.NewV4().String(), 2, db, logger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can't donate without a proper DB-loaded DonationRequest (ID is required)."))
			})

			It("Should fail if no player to donate", func() {
				game, err := GetTestGame(db, logger, true, map[string]interface{}{
					"LimitOfCardsInEachDonationRequest": 2,
				})
				Expect(err).NotTo(HaveOccurred())

				dr, err := GetTestDonationRequest(game, db, logger)

				err = dr.Donate("", 2, db, logger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Player is required to create a new Donation."))
			})

			It("Should fail if zero items donated", func() {
				game, err := GetTestGame(db, logger, true, map[string]interface{}{
					"LimitOfCardsInEachDonationRequest": 2,
				})
				Expect(err).NotTo(HaveOccurred())

				dr, err := GetTestDonationRequest(game, db, logger)

				err = dr.Donate(uuid.NewV4().String(), 0, db, logger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Amount is required to create a new Donation."))
			})
		})

		Describe("Measure", func() {
			var game *models.Game
			var donationRequest *models.DonationRequest
			var err error

			BeforeOnce(func() {
				game, err = GetTestGame(db, logger, true, map[string]interface{}{
					"LimitOfCardsInEachDonationRequest": 10000,
				})
				Expect(err).NotTo(HaveOccurred())

				itemID := GetFirstItem(game).Key
				playerID := uuid.NewV4().String()
				clanID := uuid.NewV4().String()
				donationRequest = models.NewDonationRequest(
					game.ID,
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
						donorID,
						1,
						db,
						logger,
					)
					Expect(err).NotTo(HaveOccurred())
				})

				Expect(runtime.Seconds()).Should(BeNumerically("<", 0.1), "Operation shouldn't take this long.")
			}, 5)
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
				Expect(dbDonationRequest.GameID).To(Equal(game.ID))
			})

			It("Should fail if donation request does not exist", func() {
				_, err := GetTestGame(db, logger, true)
				Expect(err).NotTo(HaveOccurred())

				drID := uuid.NewV4().String()
				dbDonationRequest, err := models.GetDonationRequestByID(drID, db, logger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(
					fmt.Sprintf("Document with id %s was not found in collection donationRequest.", drID),
				))
				Expect(dbDonationRequest).To(BeNil())
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

				Expect(runtime.Seconds()).Should(BeNumerically("<", 0.1), "Operation shouldn't take this long.")
			}, 5)
		})
	})
})
