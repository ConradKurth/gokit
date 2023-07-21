package postgres

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"time"

	"github.com/ConradKurth/gokit/config"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
)

type options struct {
	name                config.Database
	runMigration        bool
	migrationsTableName string
	migrationDir        string
	driveName           string
	url                 string
}

func WithMigrationDir(d string) func(*options) {
	return func(o *options) {
		o.migrationDir = d
	}
}

func WithConfigName(n config.Database) func(*options) {
	return func(o *options) {
		o.name = n
	}
}

func WithRunMigration() func(*options) {
	return func(o *options) {
		o.runMigration = true
	}
}

func WithDriverName(n string) func(*options) {
	return func(o *options) {
		o.driveName = n
	}
}

func WithMigrationsTableName(t string) func(*options) {
	return func(o *options) {
		o.migrationsTableName = t
	}
}

func WitURL(u string) func(*options) {
	return func(o *options) {
		o.url = u
	}
}

// ONLY CALL THIS ONCE FOR EACH DB TYPE
func InitDB(c *config.Config, opts ...func(*options)) *sqlx.DB {

	d := &options{
		driveName: "pgx",
	}
	for _, o := range opts {
		o(d)
	}

	var url string
	switch d.driveName {
	case "pg_test", "ts_test":
		if d.url != "" {
			url = d.url
		}
	case "pgx":
		u := fmt.Sprintf("postgresql://%v:%v@%v:%v/%v",
			c.GetString(fmt.Sprintf("%v.user", d.name)),
			c.GetString(fmt.Sprintf("%v.password", d.name)),
			c.GetString(fmt.Sprintf("%v.host", d.name)),
			c.GetString(fmt.Sprintf("%v.port", d.name)),
			c.GetString(fmt.Sprintf("%v.dbname", d.name)),
		)
		fmt.Println(u)
		mode := c.GetString(fmt.Sprintf("%v.sslmode", d.name))
		if mode != "" {
			u = fmt.Sprintf("%v?sslmode=%v", u, mode)
		}

		config, err := pgx.ParseConfig(u)
		if err != nil {
			panic(err)
		}
		// THIS ASSUMES THE SSL Cert is in the right location
		if mode == "verify-ca" {
			roots := x509.NewCertPool()
			sslCA := c.GetString(fmt.Sprintf("%v.certificate", d.name))
			ok := roots.AppendCertsFromPEM([]byte(sslCA))
			if !ok {
				panic("failed to parse root certificate")
			}
			config.TLSConfig = &tls.Config{
				ServerName: c.GetString(fmt.Sprintf("%v.serverName", d.name)),
				RootCAs:    roots,
			}
		}
		url = stdlib.RegisterConnConfig(config)
	default:
		log.Panicf("Unsupported driver: %v", d.driveName)
	}

	db := otelsqlx.MustConnect(d.driveName, url)
	db.DB.SetConnMaxIdleTime(time.Minute)

	if d.runMigration {
		RunMigrations(db, d)
	}

	return db
}
