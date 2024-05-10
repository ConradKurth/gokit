package responses

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ConradKurth/gokit/logger"
	"github.com/gocarina/gocsv"
)

// ResponseType are supported response types we have supported from our server
type ResponseType string

const (
	// JSON is a standard json response type
	JSON ResponseType = "application/json"

	// CSV is a csv response type
	CSV ResponseType = "text/csv"
)

// ResponseOverride can be implemented by the returning data struct to override the return type
type ResponseOverride interface {
	ResponseType() ResponseType
}

// ResponseCodeOverride can be implemented by the returning data struct to override the return type
type ResponseCodeOverride interface {
	ResponseCode() int
}

// Empty will return empty responses
func Empty(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}

func responseHandler(ctx context.Context, w http.ResponseWriter, data interface{}, code int) {

	if data == nil {
		w.WriteHeader(code)
		return
	}

	contentType := JSON
	v, ok := data.(ResponseOverride) // data *might be* responseoverride interface and v is a copy of it
	if ok {
		contentType = v.ResponseType()
	}

	override, ok := data.(ResponseCodeOverride)
	if ok {
		code = override.ResponseCode()
	}

	var marshaller func(i interface{}) ([]byte, error)
	switch contentType {
	case CSV:
		marshaller = gocsv.MarshalBytes
	default:
		marshaller = json.Marshal
	}

	b, err := marshaller(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		l := logger.GetLogger(ctx)
		l.ErrorCtx(ctx, "Error marshalling response", logger.ErrField(err))
		return
	}

	w.Header().Set("Content-Type", string(contentType))
	w.WriteHeader(code)
	if _, err := w.Write(b); err != nil {
		l := logger.GetLogger(ctx)
		l.ErrorCtx(ctx, "Error writing response", logger.ErrField(err))
		return
	}

}

// Error will return an error response
func Error(ctx context.Context, w http.ResponseWriter, e error, code int) {
	responseHandler(ctx, w, e, code)
}

// ErrHandler handled errors returned from our routes and process them as needed
func ErrHandler(h func(w http.ResponseWriter, req *http.Request) error) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		err := h(w, req)
		if err == nil {
			return
		}

		if errors.Is(err, context.Canceled) {
			w.WriteHeader(499) // client terminated error code
			return
		}

		errs, ok := IsErrorResponse(err)
		// basically saying "if err is not an errorResponse struct "
		if !ok {
			// then create a new errorresponse struct
			errs = NewErrorResponse(http.StatusInternalServerError, "Unknown error returned")
		}
		Error(req.Context(), w, errs, errs.Code)
	}

}

// Success formats the response data and writes 'OK' header
func Success(ctx context.Context, w http.ResponseWriter, data interface{}) {
	responseHandler(ctx, w, data, http.StatusOK)
}

// Accepted formats the response data and writes 'Accepted' header
func Accepted(ctx context.Context, w http.ResponseWriter, data interface{}) {
	responseHandler(ctx, w, data, http.StatusAccepted)
}
