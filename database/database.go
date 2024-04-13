package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

var qb = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

var DB *Database

type Database struct {
	db *pgxpool.Pool
}

func (d *Database) GetDb() DBTX {
	return d.db
}

func (d *Database) BeginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := d.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}
	txObj := tx
	return txObj, nil
}

func NewDatabase(ctx context.Context) {

	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword("postgres", "postgres"),
		Host:   fmt.Sprintf("%s:%d", "localhost", 5432),
		Path:   "hack",
	}

	q := dsn.Query()

	q.Add("sslmode", "disable")
	q.Add("timezone", "Asia/Almaty")

	dsn.RawQuery = q.Encode()
	poolConfig, err := pgxpool.ParseConfig(dsn.String())
	if err != nil {
		log.Fatal(err)
	}

	poolConfig.MaxConns = 15
	poolConfig.MaxConnIdleTime = time.Minute * 10

	pgxPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal(err)
	}

	if err := pgxPool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	DB = &Database{db: pgxPool}
}

func Up() {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, "postgres", "postgres", "hack")
	log.Println(dsn)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		log.Printf("cannot connect to %s \n error: $v", dsn, err)
	}

	m, err := NewMigrator(db, "hack")
	if err != nil {
		log.Printf("Could not create instance of migrator: %s\n", err.Error())
		return
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Printf("up migration error error: %s\n", err)
		return
	}
	log.Printf("up migration finished.\n")
}

func NewMigrator(conn *sql.DB, name string) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		return nil, err
	}
	return migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", "migrations"),
		name,
		driver)
}
