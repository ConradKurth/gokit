package responses

import (
	"errors"
	"fmt"
)

// ErrorType is the type of error returned
type ErrorType string

const (
	// MessageError is an error with a message
	MessageError ErrorType = "messageError"
)

// ErrorResponse is our error container
// nolint: errname
type ErrorResponse struct {
	Details []error `json:"details,omitempty"`
	Code    int     `json:"code"`
	Message string  `json:"message"`
}

// NewErrorResponse will return a new error response
func NewErrorResponse(code int, message string, errs ...error) *ErrorResponse {
	c := ErrorResponse{
		Code:    code,
		Message: message,
	}
	c.AddDetails(errs...)
	return &c
}

// Error will format the error
func (c ErrorResponse) Error() string {
	output := fmt.Sprintf("Code: %v. Message %v.     ", c.Code, c.Message)
	for _, c := range c.Details {
		output += " " + c.Error()
	}
	return output
}

// AddDetails if the error is of the type error response, then just add the details to this
func (c *ErrorResponse) AddDetails(errs ...error) {
	for _, err := range errs {
		var e ErrorResponse
		if errors.As(err, &e) {
			c.AddDetails(e.Details...)
			continue
		}
		c.Details = append(c.Details, e)
	}
}

// IsErrorResponse will return if this is an error response
func IsErrorResponse(e error) (*ErrorResponse, bool) {
	var errs *ErrorResponse
	return errs, errors.As(e, &errs)
}
