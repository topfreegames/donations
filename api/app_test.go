package api_test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/donations/api"
	. "github.com/topfreegames/donations/testing"
	"github.com/uber-go/zap"
)

var _ = Describe("App", func() {
	var logger zap.Logger
	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)
	})

	Describe("App creation", func() {
		It("should create new app", func() {
			app, err := api.GetApp("127.0.0.1", 9999, GetConfPath(), true, logger, false, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(app).NotTo(BeNil())
			Expect(app.Host).To(Equal("127.0.0.1"))
			Expect(app.Port).To(Equal(9999))
			Expect(app.Debug).To(BeTrue())
			Expect(app.Config).NotTo(BeNil())
		})
	})

	Describe("App authentication", func() {
		BeforeEach(func() {
			var err error
			_, err = api.GetApp("127.0.0.1", 9999, GetConfPath(), true, logger, false, false)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Error Handler", func() {
		var sink *TestBuffer
		BeforeEach(func() {
			sink = &TestBuffer{}
			logger = zap.New(
				zap.NewJSONEncoder(zap.NoTime()),
				zap.Output(sink),
				zap.WarnLevel,
			)
		})

		It("should handle errors and send to raven", func() {
			app, err := api.GetApp("127.0.0.1", 9999, GetConfPath(), false, logger, false, false)
			Expect(err).NotTo(HaveOccurred())

			app.OnErrorHandler(fmt.Errorf("some error occurred"), []byte("stack"))
			result := sink.Buffer.String()
			var obj map[string]interface{}
			err = json.Unmarshal([]byte(result), &obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(obj["level"]).To(Equal("error"))
			Expect(obj["msg"]).To(Equal("Panic occurred."))
			Expect(obj["operation"]).To(Equal("OnErrorHandler"))
			Expect(obj["error"]).To(Equal("some error occurred"))
			Expect(obj["stack"]).To(Equal("stack"))
		})

		It("should handle errors when they are error objects", func() {
			app, err := api.GetApp("127.0.0.1", 9999, GetConfPath(), false, logger, false, false)
			Expect(err).NotTo(HaveOccurred())

			app.OnErrorHandler(fmt.Errorf("some other error occurred"), []byte("stack"))
			result := sink.Buffer.String()
			var obj map[string]interface{}
			err = json.Unmarshal([]byte(result), &obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(obj["level"]).To(Equal("error"))
			Expect(obj["msg"]).To(Equal("Panic occurred."))
			Expect(obj["operation"]).To(Equal("OnErrorHandler"))
			Expect(obj["stack"]).To(Equal("stack"))
		})
	})

	Describe("App Load Configuration", func() {
		It("Should load configuration from file", func() {
			app, err := api.GetApp("127.0.0.1", 9999, GetConfPath(), false, logger, false, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(app.Config).NotTo(BeNil())
			expected := app.Config.GetString("healthcheck.workingText")
			Expect(expected).To(Equal("WORKING"))
		})

		It("Should faild if configuration file does not exist", func() {
			app, err := api.GetApp("127.0.0.1", 9999, "../config/invalid.yaml", false, logger, false, false)
			Expect(app).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not load configuration file from: ../config/invalid.yaml"))
		})
	})
})
