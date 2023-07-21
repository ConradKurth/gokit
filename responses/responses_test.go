package responses_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/ConradKurth/gokit/responses"
	"github.com/stretchr/testify/mock"
)

type mockWriter struct {
	mock.Mock
}

func (m *mockWriter) Header() http.Header {
	return http.Header{}
}

func (m *mockWriter) Write(b []byte) (int, error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *mockWriter) WriteHeader(statusCode int) {
	m.Called(statusCode)
}

type sendData struct {
	Hello string `json:"foobar"`
}

type codeOverride struct {
	Test string `json:"test"`
}

func (c codeOverride) ResponseCode() int {
	return 419
}

type csvData []data

func (c *csvData) ResponseType() responses.ResponseType {
	return responses.CSV
}

type data struct {
	Test string `csv:"test"`
}

func Test_Success(t *testing.T) {

	tt := []struct {
		Name   string
		Data   interface{}
		Writer func() *mockWriter
	}{
		{
			Name: "Happy path with no data",
			Writer: func() *mockWriter {
				m := &mockWriter{}
				m.On("WriteHeader", http.StatusOK).Once()
				return m
			},
		},
		{
			Name: "Happy path with data in a json form",
			Data: sendData{Hello: "cow"},
			Writer: func() *mockWriter {
				m := &mockWriter{}
				m.On("WriteHeader", http.StatusOK).Once()
				m.On("Write", []byte(`{"foobar":"cow"}`)).Return(5, nil)
				return m
			},
		},
		{
			Name: "Happy path with data in a csv form",
			Data: &csvData{
				{Test: "cow"},
			},
			Writer: func() *mockWriter {
				m := &mockWriter{}
				m.On("WriteHeader", http.StatusOK).Once()
				m.On("Write", []byte(`test
cow
`)).Return(5, nil)
				return m
			},
		},
		{
			Name: "Override ocde",
			Data: &codeOverride{
				Test: "foo",
			},
			Writer: func() *mockWriter {
				m := &mockWriter{}
				m.On("WriteHeader", 419).Once()
				m.On("Write", []byte(`{"test":"foo"}`)).Return(5, nil)
				return m
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			w := tc.Writer()
			responses.Success(context.Background(), w, tc.Data)
			w.AssertExpectations(t)
		})
	}
}

func Test_Accepted(t *testing.T) {

	tt := []struct {
		Name   string
		Data   interface{}
		Writer func() *mockWriter
	}{
		{
			Name: "Happy path with no data",
			Writer: func() *mockWriter {
				m := &mockWriter{}
				m.On("WriteHeader", http.StatusAccepted).Once()
				return m
			},
		},
		{
			Name: "Happy path with data in a json form",
			Data: sendData{Hello: "cow"},
			Writer: func() *mockWriter {
				m := &mockWriter{}
				m.On("WriteHeader", http.StatusAccepted).Once()
				m.On("Write", []byte(`{"foobar":"cow"}`)).Return(5, nil)
				return m
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			w := tc.Writer()
			responses.Accepted(context.Background(), w, tc.Data)
			w.AssertExpectations(t)
		})
	}
}

func Test_Error(t *testing.T) {
	tt := []struct {
		Name   string
		Code   int
		e      error
		Writer func() *mockWriter
	}{
		{
			Name: "Happy path with no data",
			Code: 500,
			e:    errors.New("stringme"),
			Writer: func() *mockWriter {
				m := &mockWriter{}
				m.On("WriteHeader", http.StatusInternalServerError).Once()
				m.On("Write", []byte(`{}`)).Once().Return(5, nil)
				return m
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			w := tc.Writer()
			responses.Error(context.Background(), w, tc.e, tc.Code)
			w.AssertExpectations(t)
		})
	}
}

func Test_ErrHandler(t *testing.T) {
	tt := []struct {
		Name   string
		Writer func() *mockWriter
		Arg    func(http.ResponseWriter, *http.Request) error
	}{
		{
			Name: "Happy path, no error happened",
			Writer: func() *mockWriter {
				m := &mockWriter{}
				return m
			},
			Arg: func(w http.ResponseWriter, r *http.Request) error {
				return nil
			},
		},
		{
			Name: "Happy path, with a generic error",
			Writer: func() *mockWriter {
				m := &mockWriter{}
				m.On("WriteHeader", http.StatusInternalServerError).Once()
				m.On("Write", []byte(`{"code":500,"message":"Unknown error returned"}`)).Once().Return(5, nil)
				return m
			},
			Arg: func(w http.ResponseWriter, r *http.Request) error {
				return responses.NewErrorResponse(http.StatusInternalServerError, "Unknown error returned")
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			w := tc.Writer()
			f := responses.ErrHandler(tc.Arg)
			f(w, &http.Request{})
			w.AssertExpectations(t)
		})
	}
}

func Test_Empty(t *testing.T) {
	tt := []struct {
		Name   string
		Code   int
		Writer func() *mockWriter
	}{
		{
			Name: "Happy path with no data",
			Code: 500,
			Writer: func() *mockWriter {
				m := &mockWriter{}
				m.On("WriteHeader", http.StatusInternalServerError).Once()
				return m
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			w := tc.Writer()
			responses.Empty(w, tc.Code)
			w.AssertExpectations(t)
		})
	}
}
