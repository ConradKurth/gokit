package temporaltest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_MockRun_Get(t *testing.T) {
	ctx := context.Background()

	run := &MockRun{Value: "test"}
	run.On("Get", mock.Anything, mock.Anything).Once().Return(nil)

	var s string
	require.NoError(t, run.Get(ctx, &s))
	assert.Equal(t, "test", s)

	type stru struct {
		s string
	}

	run = &MockRun{Value: stru{
		s: "test",
	}}
	run.On("Get", mock.Anything, mock.Anything).Once().Return(nil)

	res := stru{}
	require.NoError(t, run.Get(ctx, &res))
	assert.Equal(t, "test", res.s)

	run = &MockRun{Value: &stru{
		s: "test",
	}}
	run.On("Get", mock.Anything, mock.Anything).Once().Return(nil)

	res = stru{}
	require.NoError(t, run.Get(ctx, &res))
	assert.Equal(t, "test", res.s)

	run = &MockRun{Value: nil}
	run.On("Get", mock.Anything, mock.Anything).Once().Return(nil)

	resPtr := &stru{}
	require.NoError(t, run.Get(ctx, &resPtr))
}
