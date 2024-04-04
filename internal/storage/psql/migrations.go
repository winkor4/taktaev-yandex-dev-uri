package psql

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS shorten_urls (
	original_url text UNIQUE,
	short_key text
);`,
}

var queryInsert =
	`INSERT INTO shorten_urls 
	(
		original_url, 
		short_key
	)
	VALUES 
	(
		$1, 
		$2
	)
	ON CONFLICT (original_url) DO NOTHING;`

var querySelectURL =
	`SELECT original_url
	FROM shorten_urls
	WHERE short_key = $1`