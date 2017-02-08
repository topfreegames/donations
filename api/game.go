package api

import (
	"net/http"

	"github.com/topfreegames/donations/errors"
	"github.com/topfreegames/donations/log"
	"github.com/topfreegames/donations/models"

	"github.com/labstack/echo"
	"github.com/uber-go/zap"
)

//UpdateGameHandler is the handler responsible for updating games
func UpdateGameHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		gameID := c.Param("gameID")
		l := app.Logger.With(
			zap.String("source", "UpdateGameHandler"),
			zap.String("operation", "UpdateGame"),
			zap.String("gameID", gameID),
		)
		c.Set("route", "UpdateGame")

		log.D(l, "Updating game...")
		var payload UpdateGamePayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			log.E(l, "Invalid json payload!", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(400, getErrorBody(err), c)
		}

		game, err := models.GetGameByID(gameID, app.MongoDb, app.Logger)
		if err != nil {
			if _, ok := err.(*errors.DocumentNotFoundError); !ok {
				log.E(l, "Failed to retrieve game!", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return FailWith(500, err.Error(), c)
			}
			log.D(l, "Game not found, creating new game...")
			game = models.NewGame(
				payload.Name, gameID,
				payload.DonationCooldownHours,
				payload.DonationRequestCooldownHours,
			)
		} else {
			game.Name = payload.Name
			game.DonationCooldownHours = payload.DonationCooldownHours
			game.DonationRequestCooldownHours = payload.DonationRequestCooldownHours
		}

		err = game.Save(app.MongoDb, app.Logger)
		if err != nil {
			log.E(l, "Failed to update game!", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}

		log.I(l, "Updated game successfully.", func(cm log.CM) {
			cm.Write(zap.String("ID", game.ID))
		})

		gameJSON, err := game.ToJSON()
		if err != nil {
			log.E(l, "Failed to marshal game!", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}

		return c.String(http.StatusOK, string(gameJSON))
	}
}
