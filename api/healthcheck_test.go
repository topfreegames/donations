// donations
// https://github.com/topfreegames/donations
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/donations/api"
	. "github.com/topfreegames/donations/testing"
	"github.com/uber-go/zap"
)

var _ = Describe("Healthcheck API Handler", func() {
	Describe("Healthcheck Handler", func() {
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

		It("Should respond with default WORKING string", func() {
			status, body := Get(app, "/healthcheck")
			Expect(status).To(Equal(http.StatusOK))
			Expect(body).To(Equal("WORKING"))
		})

		It("Should respond with customized WORKING string", func() {
			app.Config.Set("healthcheck.workingText", "OTHERWORKING")
			status, body := Get(app, "/healthcheck")
			Expect(status).To(Equal(http.StatusOK))
			Expect(body).To(Equal("OTHERWORKING"))
		})
	})
})
