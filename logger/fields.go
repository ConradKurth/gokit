package logger

type Field interface {
	Value() interface{}
	Key() string
}

const errorKey = "error"

type field struct {
	key   string
	value interface{}
}

func (f *field) Key() string {
	return f.key
}

func (f *field) Value() interface{} {
	return f.value
}

func Any(key string, value interface{}) Field {
	return &field{
		key,

		value,
	}
}

// ErrField will return an error field while also sending the error to sentry
func ErrField(err error) Field {
	return &field{
		key:   errorKey,
		value: err,
	}
}
