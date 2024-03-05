package postgres

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"time"

	"github.com/ConradKurth/gokit/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
)

type ConnectionURLFields struct {
	User     string
	Password string
	Host     string
	Port     int
	DBName   string
	Mode     string
}

type SSLCertFields struct {
	SSLCA      string
	ServerName string
}
type options struct {
	name                config.Database
	runMigration        bool
	migrationsTableName string
	migrationDir        string
	migrationUseDBName  bool
	driveName           string
	url                 string
	trace               pgx.QueryTracer
	urlFields           *ConnectionURLFields
	sslCertsField       *SSLCertFields
}

func WithMigrationUseDBName() func(*options) {
	return func(o *options) {
		o.migrationUseDBName = true
	}
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

func WithTracer(t pgx.QueryTracer) func(*options) {
	return func(o *options) {
		o.trace = t
	}
}

func WithConnectionURLField(f ConnectionURLFields) func(*options) {
	return func(o *options) {
		o.urlFields = &f
	}
}

func WithSSLCertField(f SSLCertFields) func(*options) {
	return func(o *options) {
		o.sslCertsField = &f
	}
}

// ONLY CALL THIS ONCE FOR EACH DB TYPE
func InitDB(opts ...func(*options)) *sqlx.DB {
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
		if d.urlFields != nil {
			url = fmt.Sprintf("postgresql://%v:%v@%v:%v/%v",
				d.urlFields.User,
				d.urlFields.Password,
				d.urlFields.Host,
				d.urlFields.Port,
				d.urlFields.DBName,
			)
			if d.urlFields.Mode != "" {
				url = fmt.Sprintf("%v?sslmode=%v", url, d.urlFields.Mode)
			}

		} else {
			url = d.url
		}

		if d.url == "" && d.urlFields == nil {
			panic("No URL or URLFields provided")
		}
		config, err := pgx.ParseConfig(url)
		if err != nil {
			panic(err)
		}
		if d.trace != nil {
			config.Tracer = d.trace
		}

		// THIS ASSUMES THE SSL Cert is in the right location
		if d.sslCertsField != nil {
			roots := x509.NewCertPool()
			sslCA := d.sslCertsField.SSLCA
			ok := roots.AppendCertsFromPEM([]byte(sslCA))
			if !ok {
				panic("failed to parse root certificate")
			}
			config.TLSConfig = &tls.Config{
				ServerName: d.sslCertsField.ServerName,
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
