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
}

//Error string
func (err LimitOfCardsInDonationRequestReachedError) Error() string {
	return "This donation request is already finished."
}
