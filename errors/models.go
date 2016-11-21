package errors

import "fmt"

//DocumentNotFoundError happens when a query should return a document but it's not found
type DocumentNotFoundError struct {
	Collection string
	ID         string
}

//NewDocumentNotFoundError creates a new error
func NewDocumentNotFoundError(collection string, id string) *DocumentNotFoundError {
	return &DocumentNotFoundError{
		Collection: collection,
		ID:         id,
	}
}

//Error string
func (err DocumentNotFoundError) Error() string {
	return fmt.Sprintf("Document with id %s was not found in collection %s.", err.ID, err.Collection)
}

//ParameterIsRequiredError happens when a parameter for a given instance to be saved is required but is empty
type ParameterIsRequiredError struct {
	Parameter string
	Model     string
}

//Error string
func (err ParameterIsRequiredError) Error() string {
	return fmt.Sprintf("%s is required to create a new %s.", err.Parameter, err.Model)
}

//ItemNotFoundInGameError happens when a donation happens for an item that's not in the game
type ItemNotFoundInGameError struct {
	ItemKey string
	GameID  string
}

//Error string
func (err ItemNotFoundInGameError) Error() string {
	return fmt.Sprintf("Item %s was not found in game %s.", err.ItemKey, err.GameID)
}

//LimitOfCardsInDonationRequestReachedError happens when a donation happens for an item that's not in the game
type LimitOfCardsInDonationRequestReachedError struct {
	GameID            string
	DonationRequestID string
	ItemKey           string
	Amount            int
}

//Error string
func (err LimitOfCardsInDonationRequestReachedError) Error() string {
	return "This donation request can't accept this donation."
}

//LimitOfCardsPerPlayerInDonationRequestReachedError happens when a donation happens for an item that's not in the game
type LimitOfCardsPerPlayerInDonationRequestReachedError struct {
	GameID               string
	DonationRequestID    string
	ItemKey              string
	Amount               int
	Player               string
	CurrentDonationCount int
}

//Error string
func (err LimitOfCardsPerPlayerInDonationRequestReachedError) Error() string {
	return "This donation request can't accept any more donations from this player."
}

//DonationRequestCooldownViolatedError happens when a donation request
//is created in less than <cooldown> hours after last one
type DonationRequestCooldownViolatedError struct {
	GameID  string
	ItemKey string
	Time    int64
}

//Error string
func (err DonationRequestCooldownViolatedError) Error() string {
	return "This player can't create a new donation request so soon."
}

//DonationCooldownViolatedError happens when a donation request
//is created in less than <cooldown> hours after last one
type DonationCooldownViolatedError struct {
	GameID               string
	PlayerID             string
	TotalWeightForPeriod int
	MaxWeightForPerior   int
}

//Error string
func (err DonationCooldownViolatedError) Error() string {
	return "This player can't donate so soon."
}
