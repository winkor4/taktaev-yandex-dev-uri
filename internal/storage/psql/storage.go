package psql

import (
	"context"
	"database/sql"

	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"
)

type DB struct {
	db *sql.DB
}

func New(dsn string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	out := new(DB)
	out.db = db

	ctx := context.Background()
	err = out.Ping(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, migration := range migrations {
		if _, err := tx.Exec(migration); err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, err
	}

	return &DB{db: db}, nil
}

func (db *DB) CloseDB() error {
	return db.db.Close()
}

func (db *DB) Ping(ctx context.Context) error {
	if err := db.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func (db *DB) Set(urls []model.URL) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}

	ctx := context.Background()
	for i, url := range urls {
		result, err := tx.ExecContext(ctx, queryInsert,
			url.OriginalURL,
			url.Key,
			url.UserID,
			false)
		if err != nil {
			tx.Rollback()
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return err
		}
		urls[i].Conflict = rowsAffected == 0
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (db *DB) Get(key string) (string, error) {
	row := db.db.QueryRowContext(context.Background(), querySelectURL, key)
	ourl := new(string)
	isDeleted := new(bool)
	err := row.Scan(ourl, isDeleted)
	if err != nil {
		return "", err
	}
	if *isDeleted {
		return "", model.ErrIsDeleted
	}
	return *ourl, nil
}

func (db *DB) DeleteTable() error {

	query := "DELETE FROM shorten_urls"

	tx, err := db.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(query); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (db *DB) GetByUser(user string) ([]model.KeyAndOURL, error) {
	rows, err := db.db.QueryContext(context.Background(), querySelectUsersURL, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]model.KeyAndOURL, 0)
	for rows.Next() {
		var url model.KeyAndOURL
		err := rows.Scan(&url.OriginalURL, &url.Key)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return urls, nil
}

func (db *DB) UpdateDeleteFlag(user string, keys []string) {
	tx, err := db.db.Begin()
	if err != nil {
		return
	}

	ctx := context.Background()
	for _, key := range keys {
		var err error
		switch {
		case user != "":
			_, err = tx.ExecContext(ctx, queryUpdateDeleteFlagUser,
				key,
				user)
		default:
			_, err = tx.ExecContext(ctx, queryUpdateDeleteFlag,
				key)
		}
		if err != nil {
			tx.Rollback()
			return
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return
	}
}
