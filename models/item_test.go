package models_test

import (
	mgo "gopkg.in/mgo.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/donations/models"
	. "github.com/topfreegames/donations/testing"
	"github.com/uber-go/zap"
)

var _ = Describe("Item Model", func() {
	var logger zap.Logger
	var session *mgo.Session
	var db *mgo.Database

	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)

		session, db = GetTestMongoDB()
	})

	AfterEach(func() {
		session.Close()
		session = nil
		db = nil
	})

	Describe("Item Auxiliary Methods", func() {
		Describe("Feature", func() {
			It("Should parse and serialize to json", func() {
				item := &models.Item{
					Key:      "some-key",
					Metadata: map[string]interface{}{"x": 1},
					LimitOfItemsInEachDonationRequest: 2,
					LimitOfItemsPerPlayerDonation:     3,
					WeightPerDonation:                 4,
					UpdatedAt:                         400,
				}

				r, err := item.ToJSON()
				Expect(err).NotTo(HaveOccurred())

				rr, err := models.GetItemFromJSON(r)
				Expect(err).NotTo(HaveOccurred())

				Expect(rr.Key).To(Equal(item.Key))
				Expect(int(rr.Metadata["x"].(float64))).To(Equal(1))
				Expect(rr.LimitOfItemsInEachDonationRequest).To(BeEquivalentTo(item.LimitOfItemsInEachDonationRequest))
				Expect(rr.LimitOfItemsPerPlayerDonation).To(BeEquivalentTo(item.LimitOfItemsPerPlayerDonation))
				Expect(rr.WeightPerDonation).To(BeEquivalentTo(item.WeightPerDonation))
			})
		})
	})
})
