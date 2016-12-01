package api

import (
	"net/http"

	"github.com/topfreegames/donations/log"
	"github.com/topfreegames/donations/models"

	"github.com/labstack/echo"
	"github.com/uber-go/zap"
)

//UpsertItemHandler is the handler responsible for creating/updating items
func UpsertItemHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		gameID := c.Param("gameID")
		itemKey := c.Param("itemKey")
		l := app.Logger.With(
			zap.String("source", "UpsertItemHandler"),
			zap.String("operation", "UpsertItem"),
			zap.String("game", gameID),
		)
		c.Set("route", "UpsertItem")

		log.D(l, "Upserting item...")

		var payload UpsertItemPayload
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
		var item *models.Item
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

			err = WithSegment("Item", c, func() error {
				item, err = game.AddItem(
					itemKey, payload.Metadata,
					payload.LimitOfItemsInEachDonationRequest,
					payload.LimitOfItemsPerPlayerDonation,
					payload.WeightPerDonation,
					app.MongoDb, app.Logger,
				)
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

		log.I(l, "Created/Updated item successfully.", func(cm log.CM) {
			cm.Write(zap.String("Key", item.Key))
		})

		var itemJSON []byte
		err = WithSegment("serialization", c, func() error {
			itemJSON, err = item.ToJSON()
			if err != nil {
				log.E(l, "Failed to marshal item!", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		return c.String(http.StatusOK, string(itemJSON))
	}
}
