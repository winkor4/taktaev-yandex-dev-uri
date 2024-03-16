package databaseSQL

import (
	"context"
	"database/sql"
	"fmt"
)

type PSQLDB struct {
	db *sql.DB
}

func (db PSQLDB) Close() {
	db.db.Close()
}

func (db PSQLDB) NotAvailable() bool {
	return db.db == nil
}

func (db PSQLDB) PingContext(ctx context.Context) error {
	return db.db.PingContext(ctx)
}

func CheckConn(databaseDSN string) (PSQLDB, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return PSQLDB{db: nil}, err
	}

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		return PSQLDB{db: nil}, nil
	}

	queryText :=
		`SELECT true AS exists 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_NAME = 'shorten_urls'`

	row := db.QueryRowContext(ctx, queryText)

	tableExists := new(bool)
	err = row.Scan(tableExists)
	if err != nil && err != sql.ErrNoRows {
		return PSQLDB{db: nil}, err
	}
	if !*tableExists {
		err = createTable(ctx, db, "shorten_urls")
		if err != nil {
			return PSQLDB{db: nil}, err
		}
	}
	return PSQLDB{db: db}, nil
}

func createTable(ctx context.Context, db *sql.DB, tableName string) error {
	queryText := fmt.Sprintf("CREATE TABLE %s (ORIGINAL_URL text, SHORT_URL text);", tableName)
	_, err := db.ExecContext(ctx, queryText)
	if err != nil {
		return err
	}
	return nil
}

func (db PSQLDB) Insert(shortURL string, originalURL string) error {
	if db.NotAvailable() {
		return nil
	}
	queryText :=
		`INSERT INTO shorten_urls (original_url, short_url)
		VALUES ($1, $2)`
	_, err := db.db.ExecContext(context.Background(), queryText, originalURL, shortURL)
	if err != nil {
		return err
	}
	return nil
}

func (db PSQLDB) SelectURL(shortURL string) (string, error) {
	queryText :=
		`SELECT original_url
		FROM shorten_urls
		WHERE short_url = $1`
	row := db.db.QueryRowContext(context.Background(), queryText, shortURL)
	ourl := new(string)
	err := row.Scan(ourl)
	if err != nil {
		return "", err
	}
	return *ourl, nil
}
