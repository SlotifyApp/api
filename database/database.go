package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type options struct {
	dbHost   *string
	dbName   *string
	password *string
	port     *uint
	uname    *string
}

type Option func(opts *options) error

func WithDBName(dbName string) Option {
	return func(options *options) error {
		options.dbName = &dbName
		return nil
	}
}

func WithDBHost(dbHost string) Option {
	return func(options *options) error {
		options.dbName = &dbHost
		return nil
	}
}

func WithPort(port uint) Option {
	return func(options *options) error {
		options.port = &port
		return nil
	}
}

func WithUsername(uname string) Option {
	return func(options *options) error {
		options.uname = &uname
		return nil
	}
}

func WithPassword(password string) Option {
	return func(options *options) error {
		options.password = &password
		return nil
	}
}

func openAndPingDBContext(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("db: failed to open: %s", err.Error())
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db: failed to ping db: %s", err.Error())
	}

	return db, nil
}

// then read from env variables.
func NewDatabaseWithContext(ctx context.Context, dbOpts ...Option) (*sql.DB, error) {
	var opts options
	for _, opt := range dbOpts {
		err := opt(&opts)
		if err != nil {
			return nil, err
		}
	}

	var err error
	var dbHost string
	if opts.dbHost == nil {
		dbEnv, present := os.LookupEnv("DB_HOST")
		if !present {
			return nil, errors.New("DB_HOST was not set in env vars")
		}
		dbHost = dbEnv
	} else {
		dbHost = *opts.dbHost
	}

	var dbName string
	if opts.dbName == nil {
		dbEnv, present := os.LookupEnv("DB_NAME")
		if !present {
			return nil, errors.New("DB_NAME was not set in env vars")
		}
		dbName = dbEnv
	} else {
		dbName = *opts.dbName
	}
	var port uint
	if opts.port == nil {
		portEnv, present := os.LookupEnv("DB_PORT")
		if !present {
			return nil, errors.New("DB_PORT was not set in env vars")
		}
		var parsedPort uint64
		if parsedPort, err = strconv.ParseUint(portEnv, 10, 32); err != nil {
			return nil, errors.New("failed to parse DB_PORT into unsigned int")
		}
		port = uint(parsedPort)
	} else {
		port = *opts.port
	}

	var uname string
	if opts.uname == nil {
		dbEnv, present := os.LookupEnv("DB_USERNAME")
		if !present {
			return nil, errors.New("DB_USERNAME was not set in env vars")
		}
		uname = dbEnv
	} else {
		uname = *opts.uname
	}

	var password string
	if opts.password == nil {
		dbEnv, present := os.LookupEnv("DB_PASSWORD")
		if !present {
			return nil, errors.New("DB_PASSWORD was not set in env vars")
		}
		password = dbEnv
	} else {
		password = *opts.password
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", uname, password, dbHost, port, dbName)
	var db *sql.DB
	if db, err = openAndPingDBContext(ctx, dsn); err != nil {
		return nil, fmt.Errorf("db: error opening and pinging db: %s", err.Error())
	}

	return db, nil
}
