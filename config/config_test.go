package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironment_Validate(t *testing.T) {

	tt := []struct {
		Name        string
		Input       Environment
		ErrExpected bool
	}{

		{
			Name:  "Valid development",
			Input: Development,
		},
		{
			Name:  "Valid staging",
			Input: Staging,
		},
		{
			Name:  "Valid production",
			Input: Production,
		},
		{
			Name:        "Invalid env",
			Input:       "cow",
			ErrExpected: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Input.validate()
			if tc.ErrExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_GetEnvironment(t *testing.T) {
	tt := []struct {
		Name     string
		Env      Environment
		Expected Environment
		Panic    bool
	}{
		{
			Name:     "Happy path get env",
			Env:      Production,
			Expected: Production,
		},
		{
			Name:     "Default to local environment when none is set",
			Expected: Local,
		},
		{
			Name:  "Panic when bad env is set",
			Env:   "cow",
			Panic: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			t.Setenv(GoENV, tc.Env.String())
			defer func() {
				r := recover()
				if r == nil && tc.Panic {
					assert.Fail(t, "Expected panic and no panic happened")
					return
				}

				if r != nil && !tc.Panic {
					assert.Fail(t, "Panic happened, but was not expected")
				}
			}()

			e := GetEnvironment()
			assert.Equal(t, tc.Expected, e)
		})
	}
}

func Test_IsDevelopment(t *testing.T) {
	t.Setenv(GoENV, Development.String())
	assert.True(t, IsDevelopment())
}

func Test_IsStaging(t *testing.T) {
	t.Setenv(GoENV, Staging.String())
	assert.True(t, IsStaging())
}

func Test_IsProduction(t *testing.T) {
	t.Setenv(GoENV, Production.String())
	assert.True(t, IsProduction())
}

func Test_LoadConfig(t *testing.T) {
	tt := []struct {
		Name  string
		Env   Environment
		Panic bool
	}{
		{
			Name: "Default to dev environment when none is set",
			Env:  Development,
		},
		{
			Name:  "Panic when bad env is set",
			Env:   "cow",
			Panic: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			t.Setenv(GoENV, tc.Env.String())
			defer func() {
				r := recover()
				if r == nil && tc.Panic {
					assert.Fail(t, "Expected panic and no panic happened")
					return
				}

				if r != nil && !tc.Panic {
					assert.Fail(t, "Panic happened, but was not expected", r)
				}
			}()

			c := LoadConfig(WithPath("."))
			assert.NotNil(t, c)
		})
	}
}

func TestConfig_GetValues(t *testing.T) {
	t.Setenv(GoENV, Development.String())
	c := LoadConfig(WithPath("."))

	c.GetString("foo")
	c.GetBool("foo")
	c.GetInt("foo")
}

func TestConfig_WithMap(t *testing.T) {
	cfg := map[string]interface{}{"test": "value"}
	c := LoadConfig(WithMap(cfg))

	assert.Equal(t, "value", c.GetString("test"))
}
