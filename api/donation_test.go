package api_test

import (
	"fmt"
	"net/http"
	"sync"

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

	Describe("Create Donation Request", func() {
		Describe("Feature", func() {
			It("Should respond with donation request json after creation", func() {
				game, err := GetTestGame(app.MongoDb, app.Logger, true)
				Expect(err).NotTo(HaveOccurred())

				playerID := uuid.NewV4().String()
				itemID := GetFirstItem(game).Key
				clanID := uuid.NewV4().String()
				payload := &api.CreateDonationRequestPayload{
					Player: playerID,
					Item:   itemID,
					Clan:   clanID,
				}
				jsonPayload, err := payload.ToJSON()
				Expect(err).NotTo(HaveOccurred())
				status, body := Post(
					app,
					fmt.Sprintf("/games/%s/donation-requests/", game.ID),
					string(jsonPayload),
				)
				Expect(status).To(Equal(http.StatusOK))

				dr, err := models.GetDonationRequestFromJSON([]byte(body))
				Expect(err).NotTo(HaveOccurred())
				Expect(dr.Player).To(Equal(playerID))
				Expect(dr.Item).To(Equal(itemID))
				Expect(dr.Clan).To(Equal(clanID))
				Expect(dr.GameID).To(Equal(game.ID))
			})
		})
	})

	Describe("Donate", func() {
		Describe("Feature", func() {
			It("Should respond with donation json after creation", func() {
				game, err := GetTestGame(app.MongoDb, app.Logger, true)
				Expect(err).NotTo(HaveOccurred())

				player, err := GetTestPlayer(game, app.MongoDb, app.Logger)
				Expect(err).NotTo(HaveOccurred())

				donation, err := GetTestDonationRequest(game, app.MongoDb, app.Logger)
				Expect(err).NotTo(HaveOccurred())

				playerID := player.ID
				payload := &api.DonationPayload{
					Player:             playerID,
					Amount:             1,
					MaxWeightPerPlayer: 50,
				}
				jsonPayload, err := payload.ToJSON()
				Expect(err).NotTo(HaveOccurred())
				status, body := Post(
					app,
					fmt.Sprintf("/games/%s/donation-requests/%s/", game.ID, donation.ID),
					string(jsonPayload),
				)
				Expect(status).To(Equal(http.StatusOK))
				Expect(body).To(Equal("{\"success\":true}"))
			})

			It("Should succeed if player has no clan", func() {
				game, err := GetTestGame(app.MongoDb, app.Logger, true)
				Expect(err).NotTo(HaveOccurred())

				player, err := GetTestPlayer(game, app.MongoDb, app.Logger)
				Expect(err).NotTo(HaveOccurred())

				donation, err := GetTestDonationRequest(game, app.MongoDb, app.Logger, "")
				Expect(err).NotTo(HaveOccurred())

				playerID := player.ID
				payload := &api.DonationPayload{
					Player:             playerID,
					Amount:             1,
					MaxWeightPerPlayer: 50,
				}
				jsonPayload, err := payload.ToJSON()
				Expect(err).NotTo(HaveOccurred())
				status, body := Post(
					app,
					fmt.Sprintf("/games/%s/donation-requests/%s/", game.ID, donation.ID),
					string(jsonPayload),
				)
				Expect(status).To(Equal(http.StatusOK))
				Expect(body).To(Equal("{\"success\":true}"))
			})

			It("Should not donate more than allowed", func() {
				var wg sync.WaitGroup
				results := []map[string]interface{}{}

				game, err := GetTestGame(app.MongoDb, app.Logger, true, map[string]interface{}{
					"LimitOfItemsInEachDonationRequest": 2,
				})
				Expect(err).NotTo(HaveOccurred())

				donation, err := GetTestDonationRequest(game, app.MongoDb, app.Logger)
				Expect(err).NotTo(HaveOccurred())

				playerID := uuid.NewV4().String()
				payload := &api.DonationPayload{
					Player:             playerID,
					Amount:             1,
					MaxWeightPerPlayer: 50,
				}
				jsonPayload, err := payload.ToJSON()
				Expect(err).NotTo(HaveOccurred())

				for i := 0; i < 10; i++ {
					wg.Add(1)

					go func(index int) {
						defer wg.Done()
						status, body := Post(
							app,
							fmt.Sprintf("/games/%s/donation-requests/%s/", game.ID, donation.ID),
							string(jsonPayload),
						)

						results = append(results, map[string]interface{}{
							"status": status,
							"body":   body,
						})
					}(i)
				}

				wg.Wait()

				ok := 0
				failed := 0
				for _, result := range results {
					if result["status"] == 200 {
						ok++
					} else {
						failed++
					}
				}
				Expect(ok).To(Equal(2))
				Expect(failed).To(Equal(8))
			})
		})
	})
})
