package pagination

import (
	"math"
	"net/http"
	"reflect"
	"strconv"

	"github.com/ConradKurth/gokit/responses"
)

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

const startToken = 0

type PageOptions struct {
	DefaultSize  int
	DefaultLimit int
}

// DoPagination will do pagination. The passed arg must be a slice
func DoPagination(i interface{}, pageToken, pageSize string) (interface{}, string, int, error) {
	options := PageOptions{
		DefaultSize:  10,
		DefaultLimit: 10,
	}

	return DoPaginationWithOptions(i, pageToken, pageSize, options)
}

// DoPaginationWithOptions will do pagination. The passed arg must be a slice and integer numbers
func DoPaginationWithOptions(i interface{}, pageToken, pageSize string, options PageOptions) (interface{}, string, int, error) {
	token, size, err := parseArguments(pageToken, pageSize, options.DefaultSize, options.DefaultLimit)
	if err != nil {
		return nil, "", 0, err
	}

	v := reflect.ValueOf(i)

	if token >= v.Len() {
		return reflect.MakeSlice(v.Type(), 0, 0).Interface(), "", size, nil
	}

	nextToken := ""
	if v.Len() > token+size {
		nextToken = strconv.Itoa(token + size)
	}

	v = v.Slice(token, v.Len())
	v = v.Slice(0, min(size, v.Len()))

	return v.Interface(), nextToken, size, nil
}

func parseArguments(token, size string, defaultSize, pageLimit int) (int, int, error) {
	pToken := startToken
	pSize := defaultSize

	if token != "" {
		inputToken, err := strconv.Atoi(token)
		if err != nil {
			return 0, 0, responses.NewErrorResponse(http.StatusBadRequest, "Bad page token passed")
		}

		if inputToken >= 0 && inputToken <= math.MaxInt32 {
			pToken = inputToken
		}
	}

	if size != "" {
		inputSize, err := strconv.Atoi(size)
		if err != nil {
			return 0, 0, responses.NewErrorResponse(http.StatusBadRequest, "Bad page token passed")
		}

		if inputSize >= 0 && inputSize <= pageLimit {
			pSize = inputSize
		}
	}

	return pToken, pSize, nil
}
