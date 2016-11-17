package models_test

import (
	"fmt"

	mgo "gopkg.in/mgo.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/donations/models"
	. "github.com/topfreegames/donations/testing"
	"github.com/uber-go/zap"
)

var _ = Describe("Game Model", func() {
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

	Describe("Saving a game", func() {
		Describe("Feature", func() {
			It("Should create a new game", func() {
				game := models.NewGame(
					uuid.NewV4().String(),
					uuid.NewV4().String(),
					1,
					2,
				)
				err := game.Save(db, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(game.ID).NotTo(BeEmpty())

				var dbGame *models.Game
				err = db.C("games").FindId(game.ID).One(&dbGame)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbGame.Name).To(Equal(game.Name))
				Expect(dbGame.ID).To(Equal(game.ID))
				Expect(dbGame.DonationCooldownHours).To(Equal(1))
				Expect(dbGame.DonationRequestCooldownHours).To(Equal(2))
			})

			It("Should update an existing game", func() {
				game := models.NewGame(
					uuid.NewV4().String(),
					uuid.NewV4().String(),
					1,
					2,
				)
				err := game.Save(db, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(game.ID).NotTo(BeEmpty())
				prevID := game.ID

				newName := uuid.NewV4().String()
				game.Name = newName
				game.DonationCooldownHours = 3
				game.DonationRequestCooldownHours = 4
				err = game.Save(db, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(game.ID).To(Equal(prevID))

				var dbGame *models.Game
				err = db.C("games").FindId(game.ID).One(&dbGame)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbGame.Name).To(Equal(newName))
				Expect(dbGame.ID).To(Equal(game.ID))
				Expect(dbGame.DonationCooldownHours).To(Equal(3))
				Expect(dbGame.DonationRequestCooldownHours).To(Equal(4))
			})
		})
		Describe("Measure", func() {
			var game *models.Game
			var i int
			BeforeOnce(func() {
				game = models.NewGame(
					uuid.NewV4().String(),
					uuid.NewV4().String(),
					1,
					2,
				)
				err := game.Save(db, logger)
				Expect(err).NotTo(HaveOccurred())
			})

			Measure("it should create games fast", func(b Benchmarker) {
				runtime := b.Time("runtime", func() {
					game := models.NewGame(
						uuid.NewV4().String(),
						uuid.NewV4().String(),
						1,
						2,
					)
					err := game.Save(db, logger)
					Expect(err).NotTo(HaveOccurred())
				})

				Expect(runtime.Seconds()).Should(BeNumerically("<", 0.5), "Operation shouldn't take this long.")
			}, 500)

			Measure("it should update games fast", func(b Benchmarker) {
				i++
				runtime := b.Time("runtime", func() {
					game.Name = fmt.Sprintf("game-%d", i)
					err := game.Save(db, logger)
					Expect(err).NotTo(HaveOccurred())
				})

				Expect(runtime.Seconds()).Should(BeNumerically("<", 0.5), "Operation shouldn't take this long.")
			}, 500)
		})
	})

	Describe("Getting a game by id", func() {
		Describe("Feature", func() {
			It("Should get an existent game by id", func() {
				game, err := GetTestGame(db, logger, true)
				Expect(err).NotTo(HaveOccurred())

				dbGame, err := models.GetGameByID(game.ID, db, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbGame.ID).To(Equal(game.ID))
				Expect(dbGame.Name).To(Equal(game.Name))
			})

			It("Should not get an unexistent game by public id", func() {
				gameID := uuid.NewV4().String()
				_, err := models.GetGameByID(gameID, db, logger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf("Document with id %s was not found in collection games.", gameID)))
			})
		})

		Describe("Measure", func() {
			var game *models.Game
			var i int
			BeforeOnce(func() {
				game = models.NewGame(
					uuid.NewV4().String(),
					uuid.NewV4().String(),
					1,
					2,
				)
				err := game.Save(db, logger)
				Expect(err).NotTo(HaveOccurred())
			})

			Measure("it should retrieve games fast", func(b Benchmarker) {
				i++
				runtime := b.Time("runtime", func() {
					_, err := models.GetGameByID(game.ID, db, logger)
					Expect(err).NotTo(HaveOccurred())
				})

				Expect(runtime.Seconds()).Should(BeNumerically("<", 0.5), "Operation shouldn't take this long.")
			}, 200)
		})
	})

	Describe("Can add Items", func() {
		Describe("Feature", func() {
			It("Should add an item to a game", func() {
				game, err := GetTestGame(db, logger, false)
				Expect(err).NotTo(HaveOccurred())

				key := uuid.NewV4().String()
				meta := map[string]interface{}{"x": 1}
				item, err := game.AddItem(key, meta, 1, 2, 3, db, logger)
				Expect(err).NotTo(HaveOccurred())

				Expect(item.Key).To(Equal(key))
				Expect(item.Metadata).To(BeEquivalentTo(meta))

				dbGame, err := models.GetGameByID(game.ID, db, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbGame.ID).To(Equal(game.ID))

				Expect(dbGame.Items).To(HaveLen(1))

				dbItem := dbGame.Items[key]
				Expect(dbItem.Key).To(Equal(key))
				Expect(dbItem.Metadata).To(BeEquivalentTo(meta))
				Expect(dbItem.WeightPerDonation).To(Equal(3))
				Expect(dbItem.LimitOfCardsInEachDonationRequest).To(Equal(1))
				Expect(dbItem.LimitOfCardsPerPlayerDonation).To(Equal(2))
			})

			It("Should add multiple items to a game", func() {
				game, err := GetTestGame(db, logger, false)
				Expect(err).NotTo(HaveOccurred())

				for i := 0; i < 10; i++ {
					key := fmt.Sprintf("multiple-keys-test-%d", i)
					meta := map[string]interface{}{"x": i}
					item, err := game.AddItem(key, meta, 1, 2, 3, db, logger)
					Expect(err).NotTo(HaveOccurred())
					Expect(item).NotTo(BeNil())
				}

				dbGame, err := models.GetGameByID(game.ID, db, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbGame.ID).To(Equal(game.ID))

				Expect(dbGame.Items).To(HaveLen(10))
				for i := 0; i < 10; i++ {
					key := fmt.Sprintf("multiple-keys-test-%d", i)
					meta := map[string]interface{}{"x": i}
					dbItem := dbGame.Items[key]
					Expect(dbItem.Metadata).To(BeEquivalentTo(meta))
					Expect(dbItem.WeightPerDonation).To(Equal(3))
					Expect(dbItem.LimitOfCardsInEachDonationRequest).To(Equal(1))
					Expect(dbItem.LimitOfCardsPerPlayerDonation).To(Equal(2))
				}
			})

			It("Should overwrite item in a game", func() {
				game, err := GetTestGame(db, logger, false)
				Expect(err).NotTo(HaveOccurred())

				key := uuid.NewV4().String()

				meta := map[string]interface{}{"x": 1}
				_, err = game.AddItem(key, meta, 1, 2, 3, db, logger)
				Expect(err).NotTo(HaveOccurred())

				meta2 := map[string]interface{}{"x": 2}
				_, err = game.AddItem(key, meta2, 4, 5, 6, db, logger)
				Expect(err).NotTo(HaveOccurred())

				dbGame, err := models.GetGameByID(game.ID, db, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbGame.ID).To(Equal(game.ID))

				dbItem := dbGame.Items[key]
				Expect(dbItem.Metadata).To(BeEquivalentTo(meta2))
				Expect(dbItem.WeightPerDonation).To(Equal(6))
				Expect(dbItem.LimitOfCardsInEachDonationRequest).To(Equal(4))
				Expect(dbItem.LimitOfCardsPerPlayerDonation).To(Equal(5))
			})
		})

		Describe("Measure", func() {
			var game *models.Game
			var err error
			var meta map[string]interface{}
			BeforeOnce(func() {
				game, err = GetTestGame(db, logger, false)
				Expect(err).NotTo(HaveOccurred())

				err = game.Save(db, logger)
				Expect(err).NotTo(HaveOccurred())

				meta = map[string]interface{}{"x": 1}
			})

			Measure("it should add items fast", func(b Benchmarker) {
				runtime := b.Time("runtime", func() {
					_, err := game.AddItem(uuid.NewV4().String(), meta, 1, 2, 3, db, logger)
					Expect(err).NotTo(HaveOccurred())
				})

				Expect(runtime.Seconds()).Should(BeNumerically("<", 0.5), "Operation shouldn't take this long.")
			}, 500)
		})
	})
})
