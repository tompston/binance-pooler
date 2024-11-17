package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	conn *pgxpool.Pool
}

func New(connStr string) (*DB, error) {
	conf, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("pg.New: failed to parse the connection string: %v", err)
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("pg.New: failed to connect to the database: %v", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("pg.New: failed to ping the database: %v", err)
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Close()              { db.conn.Close() }
func (db *DB) Conn() *pgxpool.Pool { return db.conn }
