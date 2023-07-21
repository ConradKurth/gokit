package config

type Database string

func (d Database) String() string {
	return string(d)
}
