package storetester

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ConradKurth/gokit/config"
	"github.com/ConradKurth/gokit/databases/postgres"
	"github.com/DATA-DOG/go-txdb"
	"github.com/jmoiron/sqlx"
	"github.com/romanyx/polluter"
)

const PgDriver = "pg_test"
const TimescaleDriver = "ts_test"

var connCounter int64 = 0

var once sync.Once

func register(driver, dsn string) {
	txdb.Register(driver, "pgx", dsn)
	sqlx.BindDriver(driver, sqlx.DOLLAR)
}

func LoadPgOptsFromConfigPath(path string, dbName config.Database) []func(o *option) {
	c := config.LoadConfig(config.WithPath(path))
	return LoadPGOptsFromConfig(c, dbName)
}

func LoadPGOptsFromConfig(c *config.Config, dbName config.Database) []func(o *option) {
	return []func(o *option){
		WithPostgresDBName(c.GetString(fmt.Sprintf("%v.dbname", dbName))),
		WithPgPort(c.GetString(fmt.Sprintf("%v.port", dbName))),
		WithDB(dbName),
	}
}

func WithDriver(d string) func(o *option) {
	return func(o *option) {
		o.driver = d
	}
}

func WithDB(d config.Database) func(o *option) {
	return func(o *option) {
		o.database = d
	}
}

func WithPgUrl(u string) func(o *option) {
	return func(o *option) {
		o.pgUrl = u
	}
}

func WithPgPort(p string) func(o *option) {
	return func(o *option) {
		o.pgPort = p
	}
}

func WithTsUrl(u string) func(o *option) {
	return func(o *option) {
		o.tsUrl = u
	}
}

func WithTsPort(p string) func(o *option) {
	return func(o *option) {
		o.tsPort = p
	}
}

func WithTimescaleDBName(d string) func(o *option) {
	return func(o *option) {
		o.tsDBName = d
	}
}

func WithPostgresDBName(d string) func(o *option) {
	return func(o *option) {
		o.postgresDBName = d
	}
}

type option struct {
	driver         string
	database       config.Database
	pgUrl          string
	pgPort         string
	tsUrl          string
	tsPort         string
	tsDBName       string
	postgresDBName string
}

func Setup(t *testing.T, input string, opts ...func(*option)) *sqlx.DB {
	t.Helper()

	var tsUrl string
	tsPort := "5433"
	if _, ok := os.LookupEnv("CI"); ok {
		// This is the name of the docker image we are connecting to in our circle file
		tsPort = "5432"
		tsUrl = `timescale:%v/%v`
	} else {
		tsUrl = `127.0.0.1:%v/%v`
	}

	op := option{
		driver:         PgDriver,
		pgUrl:          `127.0.0.1:%v/%v`,
		pgPort:         "5432",
		tsUrl:          tsUrl,
		tsPort:         tsPort,
		tsDBName:       "",
		postgresDBName: "",
	}

	for _, o := range opts {
		o(&op)
	}

	op.pgUrl = fmt.Sprintf(op.pgUrl, op.pgPort, op.postgresDBName)
	op.tsUrl = fmt.Sprintf(op.tsUrl, op.tsPort, op.tsDBName)

	once.Do(func() {
		register(PgDriver, `postgresql://user:password@`+op.pgUrl+`?sslmode=disable`)
		register(TimescaleDriver, `postgresql://user:password@`+op.tsUrl+`?sslmode=disable`)
	})

	out := atomic.AddInt64(&connCounter, 1)
	db := postgres.InitDB(nil,
		postgres.WithDriverName(op.driver),
		postgres.WitURL(strconv.Itoa(int(out))),
	)
	if input == "" {
		return db
	}

	p := polluter.New(polluter.PostgresEngine(db.DB))

	if err := p.Pollute(strings.NewReader(input)); err != nil {
		t.Fatalf("failed to pollute: %s", err)
	}
	return db
}
