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
					map[string]interface{}{"x": 1},
				)
				err := game.Save(db, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(game.ID).NotTo(BeEmpty())

				var dbGame *models.Game
				err = db.C("games").FindId(game.ID).One(&dbGame)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbGame.Name).To(Equal(game.Name))
				Expect(dbGame.ID).To(Equal(game.ID))
				Expect(dbGame.Options).To(BeEquivalentTo(game.Options))
			})

			It("Should update an existing game", func() {
				game := models.NewGame(
					uuid.NewV4().String(),
					uuid.NewV4().String(),
					map[string]interface{}{"x": 1},
				)
				err := game.Save(db, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(game.ID).NotTo(BeEmpty())
				prevID := game.ID

				newName := uuid.NewV4().String()
				game.Name = newName
				err = game.Save(db, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(game.ID).To(Equal(prevID))

				var dbGame *models.Game
				err = db.C("games").FindId(game.ID).One(&dbGame)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbGame.Name).To(Equal(newName))
				Expect(dbGame.ID).To(Equal(game.ID))
				Expect(dbGame.Options).To(BeEquivalentTo(game.Options))
			})
		})
		Describe("Measure", func() {
			var game *models.Game
			var i int
			BeforeOnce(func() {
				game = models.NewGame(
					uuid.NewV4().String(),
					uuid.NewV4().String(),
					map[string]interface{}{
						"a": 1,
						"b": 1,
						"c": 1,
						"d": 1,
						"e": 1,
						"f": 1,
						"g": 1,
						"h": 1,
						"i": 1,
						"j": 1,
						"k": 1,
					},
				)
				err := game.Save(db, logger)
				Expect(err).NotTo(HaveOccurred())
			})

			Measure("it should create games fast", func(b Benchmarker) {
				runtime := b.Time("runtime", func() {
					game := models.NewGame(
						uuid.NewV4().String(),
						uuid.NewV4().String(),
						map[string]interface{}{"x": 1},
					)
					err := game.Save(db, logger)
					Expect(err).NotTo(HaveOccurred())
				})

				Expect(runtime.Seconds()).Should(BeNumerically("<", 0.05), "Operation shouldn't take this long.")
			}, 200)

			Measure("it should update games fast", func(b Benchmarker) {
				i++
				runtime := b.Time("runtime", func() {
					game.Options["x"] = i
					err := game.Save(db, logger)
					Expect(err).NotTo(HaveOccurred())
				})

				Expect(runtime.Seconds()).Should(BeNumerically("<", 0.05), "Operation shouldn't take this long.")
			}, 200)
		})
	})

	Describe("Getting a game by id", func() {
		Describe("Feature", func() {
			It("Should get an existent game by id", func() {
				game, err := GetTestGame(db, logger)
				Expect(err).NotTo(HaveOccurred())

				dbGame, err := models.GetGameByID(game.ID, db, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbGame.ID).To(Equal(game.ID))
				Expect(dbGame.Name).To(Equal(game.Name))
				Expect(dbGame.Options).To(BeEquivalentTo(game.Options))
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
					map[string]interface{}{
						"a": 1,
						"b": 1,
						"c": 1,
						"d": 1,
						"e": 1,
						"f": 1,
						"g": 1,
						"h": 1,
						"i": 1,
						"j": 1,
						"k": 1,
					},
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

				Expect(runtime.Seconds()).Should(BeNumerically("<", 0.05), "Operation shouldn't take this long.")
			}, 200)
		})
	})
})
