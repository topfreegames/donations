package errors

import (
	"encoding/json"
	"fmt"
	"strings"
)

//InvalidPayloadError happens whenever a payload deserialize fails validation
type InvalidPayloadError struct {
	FailedValidations []string
}

//Error string
func (err InvalidPayloadError) Error() string {
	return fmt.Sprintf(
		"The operation failed due to: %s",
		strings.Join(err.FailedValidations, ", "),
	)
}

//JSON string
func (err InvalidPayloadError) JSON() []byte {
	data, mErr := json.Marshal(map[string]interface{}{
		"code":        "DO-001",
		"error":       "invalidPayloadError",
		"description": err.Error(),
	})
	if mErr != nil {
		return nil
	}
	return data
}
