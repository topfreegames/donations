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

	Describe("Create Item", func() {
		Describe("Feature", func() {
			It("Should respond with item json after creation", func() {
				game, err := GetTestGame(app.MongoDb, app.Logger, true)
				Expect(err).NotTo(HaveOccurred())

				itemID := uuid.NewV4().String()
				meta := map[string]interface{}{"x": 1}
				payload := &api.UpsertItemPayload{
					Metadata:                          meta,
					WeightPerDonation:                 1,
					LimitOfCardsPerPlayerDonation:     2,
					LimitOfCardsInEachDonationRequest: 3,
				}
				jsonPayload, err := payload.ToJSON()
				Expect(err).NotTo(HaveOccurred())
				status, body := Put(
					app,
					fmt.Sprintf("/games/%s/items/%s", game.ID, itemID),
					string(jsonPayload),
				)
				Expect(status).To(Equal(http.StatusOK))

				item, err := models.GetItemFromJSON([]byte(body))
				Expect(err).NotTo(HaveOccurred())
				Expect(item.Key).To(Equal(itemID))
				Expect(item.Metadata["x"]).To(Equal(float64(1)))
				Expect(item.WeightPerDonation).To(Equal(1))
				Expect(item.LimitOfCardsPerPlayerDonation).To(Equal(2))
				Expect(item.LimitOfCardsInEachDonationRequest).To(Equal(3))
			})
		})
	})
})
