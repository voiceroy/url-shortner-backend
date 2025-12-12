package store

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrorDBOffline = errors.New("database is offline")
)

type DBPool struct {
	DB *pgxpool.Pool
}

var db *DBPool

func ConnectToDB() {
	New()
}

func New() *DBPool {
	if db == nil {
		db = &DBPool{}
		dbPool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Printf("failed to open db %s", err)
			os.Exit(1)
		}

		db.DB = dbPool

		if err := db.PingContext(context.Background()); err != nil {
			log.Fatalln(err)
		}

		if err := Migrate(); !errors.Is(err, migrate.ErrNoChange) && err != nil {
			log.Fatalln(err)
		}
	}

	return db
}

func (pool *DBPool) GetConnection(c context.Context) (*pgxpool.Conn, error) {
	return pool.DB.Acquire(c)
}

func Migrate() error {
	p := &pgx.Postgres{}

	dB, err := p.Open(os.Getenv("PGX_URL"))
	if err != nil {
		log.Println(err)
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"pgx",
		dB,
	)

	if err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		return err
	}

	return nil
}

func (d *DBPool) PingContext(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := d.DB.Ping(ctx); err != nil {
		log.Println(ErrorDBOffline)
		return fmt.Errorf("DB offline: %w", err)
	}

	log.Println("DB is up and running")
	return nil
}
