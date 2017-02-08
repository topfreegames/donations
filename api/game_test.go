package api_test

import (
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

	Describe("Update Game", func() {
		It("Should update game if it exists", func() {
			game, err := GetTestGame(app.MongoDb, app.Logger, true)
			Expect(err).NotTo(HaveOccurred())
			gameName := uuid.NewV4().String()
			payload := &api.UpdateGamePayload{
				Name: gameName,
				DonationCooldownHours:        1,
				DonationRequestCooldownHours: 2,
			}
			jsonPayload, err := payload.ToJSON()
			Expect(err).NotTo(HaveOccurred())
			status, body := Put(app, fmt.Sprintf("/games/%s", game.ID), string(jsonPayload))
			Expect(status).To(Equal(http.StatusOK), body)

			rGame, err := models.GetGameFromJSON([]byte(body))
			Expect(err).NotTo(HaveOccurred())
			Expect(rGame.ID).To(Equal(game.ID))
			Expect(rGame.Name).To(Equal(gameName))
			Expect(rGame.DonationCooldownHours).To(Equal(1))
			Expect(rGame.DonationRequestCooldownHours).To(Equal(2))
		})

		It("Should create game if it does not exist", func() {
			id := uuid.NewV4().String()
			gameName := uuid.NewV4().String()
			payload := &api.UpdateGamePayload{
				Name: gameName,
				DonationCooldownHours:        1,
				DonationRequestCooldownHours: 2,
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

		FIt("Should fail with payload if error", func() {
			id := uuid.NewV4().String()
			gameName := uuid.NewV4().String()
			payload := &api.UpdateGamePayload{
				Name: gameName,
			}
			jsonPayload, err := payload.ToJSON()
			Expect(err).NotTo(HaveOccurred())
			status, body := Put(app, fmt.Sprintf("/games/%s", id), string(jsonPayload))
			Expect(status).To(Equal(http.StatusBadRequest), body)

			Expect(body).To(Equal("qwe"))
		})
	})
})
