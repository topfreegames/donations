package api

import (
	"net/http"

	"github.com/topfreegames/donations/errors"
	"github.com/topfreegames/donations/log"
	"github.com/topfreegames/donations/models"

	"github.com/labstack/echo"
	"github.com/uber-go/zap"
)

//CreateGameHandler is the handler responsible for creating new games
func CreateGameHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		l := app.Logger.With(
			zap.String("source", "CreateGameHandler"),
			zap.String("operation", "CreateGame"),
		)
		c.Set("route", "CreateGame")

		log.D(l, "Creating game...")

		var payload CreateGamePayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			log.E(l, "Invalid json payload!", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(400, err.Error(), c)
		}

		_, err := models.GetGameByID(payload.ID, app.MongoDb, app.Logger)
		if err != nil {
			if _, ok := err.(*errors.DocumentNotFoundError); !ok {
				log.E(l, "Failed to retrieve game!", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return FailWith(500, err.Error(), c)
			}
		} else {
			msg := "There is already a game with the same ID."
			log.E(l, msg, func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(409, msg, c)
		}

		game := models.NewGame(payload.Name, payload.ID, payload.Options)
		err = game.Save(app.MongoDb, app.Logger)
		if err != nil {
			log.E(l, "Failed to save game!", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}

		log.I(l, "Created new game successfully.", func(cm log.CM) {
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
			return FailWith(400, err.Error(), c)
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
			game = models.NewGame(payload.Name, gameID, payload.Options)
		} else {
			game.Name = payload.Name
			game.Options = payload.Options
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
