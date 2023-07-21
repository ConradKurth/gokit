package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DoPagination(t *testing.T) {
	tt := []struct {
		Name           string
		Input          []int
		InputToken     string
		InputSize      string
		Output         []int
		OutputToken    string
		OutputPageSize int
		ErrExpected    bool
	}{
		{
			Name:           "Single page no token & page size",
			Input:          []int{1, 2, 3, 4, 5, 6, 7, 8},
			InputToken:     "",
			InputSize:      "",
			Output:         []int{1, 2, 3, 4, 5, 6, 7, 8},
			OutputToken:    "",
			OutputPageSize: 10,
			ErrExpected:    false,
		},
		{
			Name:           "First page, no token & page size",
			Input:          []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			InputToken:     "",
			InputSize:      "",
			Output:         []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			OutputToken:    "10",
			OutputPageSize: 10,
			ErrExpected:    false,
		},
		{
			Name:           "Next page, no page size",
			Input:          []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			InputToken:     "10",
			InputSize:      "",
			Output:         []int{11, 12, 13, 14, 15, 16},
			OutputToken:    "",
			OutputPageSize: 10,
			ErrExpected:    false,
		},
		{
			Name:           "Next page",
			Input:          []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			InputToken:     "10",
			InputSize:      "3",
			Output:         []int{11, 12, 13},
			OutputToken:    "13",
			OutputPageSize: 3,
			ErrExpected:    false,
		},
		{
			Name:           "Next page with page size > limit",
			Input:          []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			InputToken:     "",
			InputSize:      "13",
			Output:         []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			OutputToken:    "10",
			OutputPageSize: 10,
			ErrExpected:    false,
		},
		{
			Name:           "Max Page Token",
			Input:          []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			InputToken:     "2147483648", // max int32 + 1
			InputSize:      "",
			Output:         nil,
			OutputToken:    "10",
			OutputPageSize: 10,
			ErrExpected:    true,
		},
		{
			Name:           "Bad Token",
			Input:          []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			InputToken:     "bad",
			InputSize:      "",
			Output:         nil,
			OutputToken:    "",
			OutputPageSize: 0,
			ErrExpected:    true,
		},
		{
			Name:           "Bad Page Size",
			Input:          []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			InputToken:     "",
			InputSize:      "bad",
			Output:         nil,
			OutputToken:    "",
			OutputPageSize: 0,
			ErrExpected:    true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			page, nextToken, pageSize, err := DoPagination(tc.Input, tc.InputToken, tc.InputSize)
			if !tc.ErrExpected {
				assert.Nil(t, err)
				assert.Equal(t, tc.Output, page)
			} else {
				assert.Nil(t, tc.Output)
			}
			assert.Equal(t, tc.OutputToken, nextToken)
			assert.Equal(t, tc.OutputPageSize, pageSize)
		})
	}
}

func Test_DoPaginationWithOptions(t *testing.T) {
	tt := []struct {
		Name           string
		Input          []int
		InputToken     string
		InputSize      string
		Options        PageOptions
		Output         []int
		OutputToken    string
		OutputPageSize int
		ErrExpected    bool
	}{
		{
			Name:           "Single page no token & page size",
			Input:          []int{1, 2, 3},
			InputToken:     "",
			InputSize:      "",
			Options:        PageOptions{DefaultSize: 3, DefaultLimit: 5},
			Output:         []int{1, 2, 3},
			OutputToken:    "",
			OutputPageSize: 3,
			ErrExpected:    false,
		},
		{
			Name:           "First page, no token & page size",
			Input:          []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			InputToken:     "",
			InputSize:      "",
			Options:        PageOptions{DefaultSize: 6, DefaultLimit: 10},
			Output:         []int{1, 2, 3, 4, 5, 6},
			OutputToken:    "6",
			OutputPageSize: 6,
			ErrExpected:    false,
		},
		{
			Name:           "Next page, no page size",
			Input:          []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			InputToken:     "10",
			InputSize:      "",
			Options:        PageOptions{DefaultSize: 9, DefaultLimit: 9},
			Output:         []int{11, 12, 13, 14, 15, 16},
			OutputToken:    "",
			OutputPageSize: 9,
			ErrExpected:    false,
		},
		{
			Name:           "Next page",
			Input:          []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			InputToken:     "10",
			InputSize:      "3",
			Options:        PageOptions{DefaultSize: 9, DefaultLimit: 9},
			Output:         []int{11, 12, 13},
			OutputToken:    "13",
			OutputPageSize: 3,
			ErrExpected:    false,
		},
		{
			Name:           "Max limit passed",
			Input:          []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			InputToken:     "0",
			InputSize:      "10",
			Options:        PageOptions{DefaultSize: 9, DefaultLimit: 10},
			Output:         []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			OutputToken:    "10",
			OutputPageSize: 10,
			ErrExpected:    false,
		},
		{
			Name:           "Next page with page size > limit",
			Input:          []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			InputToken:     "",
			InputSize:      "13",
			Options:        PageOptions{DefaultSize: 9, DefaultLimit: 9},
			Output:         []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
			OutputToken:    "9",
			OutputPageSize: 9,
			ErrExpected:    false,
		},
		{
			Name:           "Bad Token",
			Input:          []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			InputToken:     "bad",
			InputSize:      "",
			Options:        PageOptions{DefaultSize: 16, DefaultLimit: 16},
			Output:         nil,
			OutputToken:    "",
			OutputPageSize: 0,
			ErrExpected:    true,
		},
		{
			Name:           "Bad Page Size",
			Input:          []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			InputToken:     "",
			InputSize:      "bad",
			Options:        PageOptions{DefaultSize: 16, DefaultLimit: 16},
			Output:         nil,
			OutputToken:    "",
			OutputPageSize: 0,
			ErrExpected:    true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			page, nextToken, pageSize, err := DoPaginationWithOptions(tc.Input, tc.InputToken, tc.InputSize, tc.Options)
			if !tc.ErrExpected {
				assert.Nil(t, err)
				assert.Equal(t, tc.Output, page)
			} else {
				assert.Nil(t, tc.Output)
			}

			assert.Equal(t, tc.OutputToken, nextToken)
			assert.Equal(t, tc.OutputPageSize, pageSize)
		})
	}
}
