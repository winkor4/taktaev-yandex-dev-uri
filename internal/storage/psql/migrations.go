package psql

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS shorten_urls (
	original_url text UNIQUE,
	short_key text,
	user_id text,
	is_deleted bool NOT NULL
);`,
}

var queryInsert = `INSERT INTO shorten_urls 
	(
		original_url, 
		short_key,
		user_id,
		is_deleted
	)
	VALUES 
	(
		$1, 
		$2,
		$3,
		$4
	)
	ON CONFLICT (original_url) DO NOTHING;`

var querySelectURL = `SELECT 
		original_url,
		is_deleted
	FROM shorten_urls
	WHERE short_key = $1`

var querySelectUsersURL = `SELECT 
		original_url,
		short_key
	FROM shorten_urls
	WHERE 
		user_id = $1
		AND not is_deleted`

var queryUpdateDeleteFlagUser = `UPDATE shorten_urls
	SET
		is_deleted = true
	WHERE
		short_key = $1
		AND user_id = $2`

var queryUpdateDeleteFlag = `UPDATE shorten_urls
	SET
		is_deleted = true
	WHERE
		short_key = $1`
