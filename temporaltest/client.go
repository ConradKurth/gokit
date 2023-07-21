package temporaltest

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

type MockClient struct {
	mock.Mock
	client.Client
}

func (m *MockClient) SignalWithStartWorkflow(ctx context.Context, workflowID, signalName string, signalArg interface{}, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	var arg mock.Arguments
	if len(args) > 0 {
		arg = m.Called(ctx, workflowID, signalName, signalArg, options, workflow, args)
	} else {
		arg = m.Called(ctx, workflowID, signalName, signalArg, options, workflow)
	}
	return arg.Get(0).(client.WorkflowRun), arg.Error(1)
}

func (m *MockClient) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	var arg mock.Arguments
	if len(args) > 0 {
		arg = m.Called(ctx, options, workflow, args)
	} else {
		arg = m.Called(ctx, options, workflow)
	}
	return arg.Get(0).(client.WorkflowRun), arg.Error(1)
}

func (m *MockClient) SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, arg interface{}) error {
	a := m.Called(ctx, workflowID, runID, signalName, arg)
	return a.Error(0)
}

func (m *MockClient) QueryWorkflow(ctx context.Context, workflowID string, runID string, queryType string, args ...interface{}) (converter.EncodedValue, error) {
	a := m.Called(ctx, workflowID, runID, queryType, args)
	return a.Get(0).(converter.EncodedValue), a.Error(1)
}
