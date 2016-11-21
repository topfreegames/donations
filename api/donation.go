package api

import (
	"fmt"
	"net/http"

	"github.com/topfreegames/donations/log"
	"github.com/topfreegames/donations/models"

	"github.com/labstack/echo"
	"github.com/uber-go/zap"
)

//CreateDonationRequestHandler is the handler responsible for creating donation requests
func CreateDonationRequestHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		l := app.Logger.With(
			zap.String("source", "CreateDonationRequestHandler"),
			zap.String("operation", "CreateDonationRequest"),
		)
		c.Set("route", "CreateDonationRequest")
		gameID := c.Param("gameID")

		log.D(l, "Creating donation request...")

		var payload CreateDonationRequestPayload
		err := WithSegment("payload", c, func() error {
			if err := LoadJSONPayload(&payload, c, l); err != nil {
				log.E(l, "Invalid json payload!", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}

			return nil
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		var status int
		var donationRequest *models.DonationRequest
		var game *models.Game
		err = WithSegment("model", c, func() error {
			err = WithSegment("Game", c, func() error {
				game, err = models.GetGameByID(gameID, app.MongoDb, app.Logger)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}

			err = WithSegment("DonationRequest", c, func() error {
				mutexID := fmt.Sprintf("DonationRequest-%s-%s", gameID, payload.Player)
				mutex := app.GetMutex(mutexID, 32, 8) // Number of retries to get lock and Lock expiration
				err := mutex.Lock()
				if err != nil {
					log.E(l, "Could not acquire lock after many retries.", func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
					return err
				}
				defer mutex.Unlock()

				donationRequest = models.NewDonationRequest(
					game.ID,
					payload.Item,
					payload.Player,
					payload.Clan,
				)
				err = donationRequest.Create(app.MongoDb, app.Logger)
				if err != nil {
					status = 500
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(status, err.Error(), c)
		}

		log.I(l, "Created new donation request successfully.", func(cm log.CM) {
			cm.Write(zap.String("ID", donationRequest.ID))
		})

		var donationRequestJSON []byte
		err = WithSegment("serialization", c, func() error {
			donationRequestJSON, err = donationRequest.ToJSON()
			if err != nil {
				log.E(l, "Failed to marshal donation request!", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		return c.String(http.StatusOK, string(donationRequestJSON))
	}
}

//CreateDonationHandler is the handler responsible for creating donation requests
func CreateDonationHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		l := app.Logger.With(
			zap.String("source", "CreateDonationHandler"),
			zap.String("operation", "CreateDonation"),
		)
		c.Set("route", "CreateDonation")
		gameID := c.Param("gameID")
		donationRequestID := c.Param("donationRequestID")

		log.D(l, "Creating donation...")

		var payload DonationPayload
		err := WithSegment("payload", c, func() error {
			if err := LoadJSONPayload(&payload, c, l); err != nil {
				log.E(l, "Invalid json payload!", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}

			return nil
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		var donationRequest *models.DonationRequest
		err = WithSegment("model", c, func() error {
			mutexID := fmt.Sprintf("Donate-%s-%s", gameID, donationRequestID)
			mutex := app.GetMutex(mutexID, 32, 8) // Number of retries to get lock and Lock expiration
			err := mutex.Lock()
			if err != nil {
				log.E(l, "Could not acquire lock after many retries.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			defer mutex.Unlock()

			donationRequest, err = models.GetDonationRequestByID(donationRequestID, app.MongoDb, app.Logger)
			if err != nil {
				return err
			}

			err = donationRequest.Donate(payload.Player, payload.Amount, payload.MaxWeightPerPlayer, app.MongoDb, app.Logger)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		log.I(l, "Created new donation request successfully.", func(cm log.CM) {
			cm.Write(zap.String("ID", donationRequest.ID))
		})

		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		return c.String(http.StatusOK, "{\"success\":true}")
	}
}
