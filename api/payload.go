//go:generate easyjson -all -no_std_marshalers $GOFILE

package api

import (
	"fmt"

	"github.com/mailru/easyjson/jwriter"
	"github.com/topfreegames/donations/log"
	"github.com/uber-go/zap"
)

//Validatable indicates that a struct can be validated
type Validatable interface {
	Validate() []string
}

//ValidatePayload for any validatable payload
func ValidatePayload(payload Validatable) []string {
	return payload.Validate()
}

//NewValidation for validating structs
func NewValidation() *Validation {
	return &Validation{
		errors: []string{},
	}
}

//Validation struct
type Validation struct {
	errors []string
}

func (v *Validation) validateRequired(name string, value interface{}) {
	if value == "" {
		v.errors = append(v.errors, fmt.Sprintf("%s is required", name))
	}
}

func (v *Validation) validateRequiredString(name, value string) {
	if value == "" {
		v.errors = append(v.errors, fmt.Sprintf("%s is required", name))
	}
}

func (v *Validation) validateRequiredInt(name string, value int) {
	if value == 0 {
		v.errors = append(v.errors, fmt.Sprintf("%s is required", name))
	}
}

func (v *Validation) validateRequiredMap(name string, value map[string]interface{}) {
	if value == nil || len(value) == 0 {
		v.errors = append(v.errors, fmt.Sprintf("%s is required", name))
	}
}

func (v *Validation) validateCustom(name string, valFunc func() []string) {
	errors := valFunc()
	if len(errors) > 0 {
		v.errors = append(v.errors, errors...)
	}
}

//Errors in validation
func (v *Validation) Errors() []string {
	return v.errors
}

func logPayloadErrors(l zap.Logger, errors []string) {
	var fields []zap.Field
	for _, err := range errors {
		fields = append(fields, zap.String("validationError", err))
	}
	log.W(l, "Payload is not valid", func(cm log.CM) {
		cm.Write(fields...)
	})
}

//UpdateGamePayload maps the payload for the Create Game route
type UpdateGamePayload struct {
	Name                         string `json:"name"`
	DonationCooldownHours        int    `json:"donationCooldownHours" bson:"donationCooldownHours"`
	DonationRequestCooldownHours int    `json:"donationRequestCooldownHours" bson:"donationRequestCooldownHours"`
}

//Validate all the required fields for updating a game
func (ugp *UpdateGamePayload) Validate() []string {
	v := NewValidation()
	v.validateRequiredString("name", ugp.Name)
	v.validateRequiredInt("donationCooldownHours", ugp.DonationCooldownHours)
	v.validateRequiredInt("donationRequestCooldownHours", ugp.DonationRequestCooldownHours)
	return v.Errors()
}

//ToJSON returns the payload as JSON
func (ugp *UpdateGamePayload) ToJSON() ([]byte, error) {
	w := jwriter.Writer{}
	ugp.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

//CreateDonationRequestPayload maps the payload for the Create Game route
type CreateDonationRequestPayload struct {
	Item   string `json:"item"`
	Player string `json:"player"`
	Clan   string `json:"clan"`
}

//Validate all the required fields for creating a game
func (cdrp *CreateDonationRequestPayload) Validate() []string {
	v := NewValidation()
	v.validateRequiredString("item", cdrp.Item)
	v.validateRequiredString("player", cdrp.Player)
	v.validateRequiredString("clan", cdrp.Clan)
	return v.Errors()
}

//ToJSON returns the payload as JSON
func (cdrp *CreateDonationRequestPayload) ToJSON() ([]byte, error) {
	w := jwriter.Writer{}
	cdrp.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

//DonationPayload maps the payload for the Create Game route
type DonationPayload struct {
	Player             string `json:"player"`
	Amount             int    `json:"amount"`
	MaxWeightPerPlayer int    `json:"maxWeightPerPlayer"`
}

//Validate all the required fields for creating a game
func (dp *DonationPayload) Validate() []string {
	v := NewValidation()
	v.validateRequiredString("player", dp.Player)
	v.validateRequiredInt("amount", dp.Amount)
	v.validateRequiredInt("maxWeightPerPlayer", dp.MaxWeightPerPlayer)
	return v.Errors()
}

//ToJSON returns the payload as JSON
func (dp *DonationPayload) ToJSON() ([]byte, error) {
	w := jwriter.Writer{}
	dp.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

//UpsertItemPayload maps the payload for the Upsert Item route
type UpsertItemPayload struct {
	Metadata                          map[string]interface{} `json:"metadata"`
	WeightPerDonation                 int                    `json:"weightPerDonation"`
	LimitOfCardsPerPlayerDonation     int                    `json:"limitOfCardsPerPlayerDonation"`
	LimitOfCardsInEachDonationRequest int                    `json:"limitOfCardsInEachDonationRequest"`
}

//Validate all the required fields for creating a game
func (uip *UpsertItemPayload) Validate() []string {
	v := NewValidation()
	v.validateRequiredMap("amount", uip.Metadata)
	v.validateRequiredInt("weightPerDonation", uip.WeightPerDonation)
	v.validateRequiredInt("limitOfCardsPerPlayerDonation", uip.LimitOfCardsPerPlayerDonation)
	v.validateRequiredInt("limitOfCardsInEachDonationRequest", uip.LimitOfCardsInEachDonationRequest)
	return v.Errors()
}

//ToJSON returns the payload as JSON
func (uip *UpsertItemPayload) ToJSON() ([]byte, error) {
	w := jwriter.Writer{}
	uip.MarshalEasyJSON(&w)
	return w.BuildBytes()
}
