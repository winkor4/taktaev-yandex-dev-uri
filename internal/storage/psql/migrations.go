package psql

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS shorten_urls (
	original_url text UNIQUE,
	short_key text,
	user_id text
);`,
}

var queryInsert =
	`INSERT INTO shorten_urls 
	(
		original_url, 
		short_key,
		user_id
	)
	VALUES 
	(
		$1, 
		$2,
		$3
	)
	ON CONFLICT (original_url) DO NOTHING;`

var querySelectURL =
	`SELECT original_url
	FROM shorten_urls
	WHERE short_key = $1`

var querySelectUsersURL =
	`SELECT 
		original_url,
		short_key
	FROM shorten_urls
	WHERE user_id = $1`	