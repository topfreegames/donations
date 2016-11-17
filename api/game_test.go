package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/topfreegames/donations/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/donations/api"
	. "github.com/topfreegames/donations/testing"
	"github.com/uber-go/zap"
)

var _ = Describe("Game Handler", func() {
	var logger zap.Logger
	var app *api.App

	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)

		app = GetDefaultTestApp(logger)
	})

	AfterEach(func() {
		app.Stop()
	})

	Describe("Create Game", func() {
		It("Should respond with game json after creation", func() {
			gameName := uuid.NewV4().String()
			id := uuid.NewV4().String()
			payload := &api.CreateGamePayload{
				Name: gameName,
				ID:   id,
			}
			jsonPayload, err := payload.ToJSON()
			Expect(err).NotTo(HaveOccurred())
			status, body := Post(app, "/games", string(jsonPayload))
			Expect(status).To(Equal(http.StatusOK))

			game, err := models.GetGameFromJSON([]byte(body))
			Expect(err).NotTo(HaveOccurred())
			Expect(game.ID).To(Equal(id))
			Expect(game.Name).To(Equal(gameName))
		})

		It("Should fail if game with same ID already exists", func() {
			game, err := GetTestGame(app.MongoDb, app.Logger)
			Expect(err).NotTo(HaveOccurred())

			gameName := uuid.NewV4().String()
			payload := &api.CreateGamePayload{
				Name: gameName,
				ID:   game.ID,
			}
			jsonPayload, err := payload.ToJSON()
			Expect(err).NotTo(HaveOccurred())
			status, body := Post(app, "/games", string(jsonPayload))
			Expect(status).To(Equal(http.StatusConflict))

			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["reason"]).To(Equal("There is already a game with the same ID."))
		})

		It("Should fail if bad payload", func() {
			status, body := Post(app, "/games", "{}")
			Expect(status).To(Equal(http.StatusBadRequest))

			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["reason"]).To(Equal("name is required, id is required"))
		})
	})

	Describe("Update Game", func() {
		It("Should update game if it exists", func() {
			game, err := GetTestGame(app.MongoDb, app.Logger)
			Expect(err).NotTo(HaveOccurred())
			gameName := uuid.NewV4().String()
			payload := &api.UpdateGamePayload{
				Name: gameName,
			}
			jsonPayload, err := payload.ToJSON()
			Expect(err).NotTo(HaveOccurred())
			status, body := Put(app, fmt.Sprintf("/games/%s", game.ID), string(jsonPayload))
			Expect(status).To(Equal(http.StatusOK), body)

			rGame, err := models.GetGameFromJSON([]byte(body))
			Expect(err).NotTo(HaveOccurred())
			Expect(rGame.ID).To(Equal(game.ID))
			Expect(rGame.Name).To(Equal(gameName))
		})

		It("Should create game if it does not exist", func() {
			id := uuid.NewV4().String()
			gameName := uuid.NewV4().String()
			payload := &api.UpdateGamePayload{
				Name: gameName,
			}
			jsonPayload, err := payload.ToJSON()
			Expect(err).NotTo(HaveOccurred())
			status, body := Put(app, fmt.Sprintf("/games/%s", id), string(jsonPayload))
			Expect(status).To(Equal(http.StatusOK), body)

			rGame, err := models.GetGameFromJSON([]byte(body))
			Expect(err).NotTo(HaveOccurred())
			Expect(rGame.ID).To(Equal(id))
			Expect(rGame.Name).To(Equal(gameName))
		})
	})
})
