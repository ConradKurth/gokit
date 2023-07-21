package temporaltest

import (
	"context"
	"reflect"

	"github.com/stretchr/testify/mock"
)

type MockRun struct {
	mock.Mock

	Value interface{}
}

func (m *MockRun) GetID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRun) GetRunID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRun) Get(ctx context.Context, valuePtr interface{}) error {
	args := m.Called(ctx, valuePtr)

	if valuePtr != nil && m.Value != nil {
		elem := reflect.ValueOf(valuePtr).Elem()
		if reflect.ValueOf(m.Value).Kind() == reflect.Ptr {
			elem.Set(reflect.ValueOf(m.Value).Elem())
		} else {
			elem.Set(reflect.ValueOf(m.Value))
		}
	}

	return args.Error(0)
}
