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
