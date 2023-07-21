// Package ptrutil provides pointer conversion utility functions.
package ptrutil

func ToPtr[T any](i T) *T {
	return &i
}

func Deref[T any](i *T) T {
	if i == nil {
		var def T
		return def
	}
	return *i
}
