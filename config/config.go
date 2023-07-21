package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

type Environment string

const (
	GoENV = "GO_ENV"

	Development Environment = "development"

	Feature Environment = "feature"

	Staging Environment = "staging"

	Production Environment = "production"
)

func (e Environment) String() string {
	return string(e)
}

func (e Environment) validate() error {
	switch e {
	case Development, Feature, Staging, Production:
		return nil
	default:
		return fmt.Errorf("Unsupported env: '%v'", e)
	}
}

func GetEnvironment() Environment {

	goEnv := Environment(os.Getenv(GoENV))
	if goEnv == "" {
		goEnv = Development
	}

	if err := goEnv.validate(); err != nil {
		panic(err)
	}
	return goEnv
}

func IsDevelopment() bool {
	e := GetEnvironment()
	return e == Development
}

func IsStaging() bool {
	e := GetEnvironment()
	return e == Staging
}

func IsFeature() bool {
	e := GetEnvironment()
	return e == Feature
}

func IsProduction() bool {
	e := GetEnvironment()
	return e == Production
}

type Config struct {
	v    *viper.Viper
	lock sync.Mutex
}

type options struct {
	Path   string
	Reader io.Reader
}

func WithPath(p string) func(*options) {
	return func(o *options) {
		o.Path = p
	}
}

func WithReader(r io.Reader) func(*options) {
	return func(o *options) {
		o.Reader = r
	}
}

// WithMap allows passing the config as a map, useful for unit tests.
func WithMap(cfg map[string]interface{}) func(*options) {
	data, err := json.Marshal(cfg)
	if err != nil {
		panic(fmt.Sprintf("encoding config map: %s", err))
	}
	reader := bytes.NewReader(data)

	return func(o *options) {
		o.Reader = reader
	}
}

func LoadConfig(opts ...func(*options)) *Config {

	o := options{
		Path:   "./config/",
		Reader: nil,
	}
	for _, opts := range opts {
		opts(&o)
	}

	v := viper.New()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetConfigName(fmt.Sprintf("%v.config", GetEnvironment()))
	v.AddConfigPath(o.Path)

	var err error
	if o.Reader != nil {
		v.SetConfigType("json")
		err = v.ReadConfig(o.Reader)
	} else {
		err = v.ReadInConfig()
	}

	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %w\n", err))
	}

	return &Config{
		v: v,
	}
}

// SetValue ONLY USE THIS IF YOU KNOW WHAT YOU ARE DOING!
func (c *Config) SetValue(k string, v interface{}) {
	c.v.Set(k, v)
}

func (c *Config) bindEnv(e string) {
	if err := c.v.BindEnv(e); err != nil {
		panic(err)
	}
}

func (c *Config) GetStringMapString(s string) map[string]string {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.bindEnv(s)
	return c.v.GetStringMapString(s)
}

func (c *Config) GetFloat64(s string) float64 {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.bindEnv(s)

	return c.v.GetFloat64(s)
}

func (c *Config) GetString(s string) string {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.bindEnv(s)

	return c.v.GetString(s)
}

func (c *Config) GetStringSlice(s string) []string {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.bindEnv(s)
	return c.v.GetStringSlice(s)
}

func (c *Config) GetBool(s string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.bindEnv(s)
	return c.v.GetBool(s)
}

func (c *Config) GetBoolDefault(s string, d bool) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.bindEnv(s)

	if v := c.v.Get(s); v == nil {
		return d
	}

	return c.v.GetBool(s)
}

func (c *Config) GetInt(s string) int {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.bindEnv(s)
	return c.v.GetInt(s)
}
