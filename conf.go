package main

// Config config
type Config struct {
	Host string
	Port int
	DB DBInfo
	API []APIInfo
}

// DBInfo db config
type DBInfo struct {
	URL 		string	`toml:"url"`
	MaxPoolSize int		`toml:"max_pool_size"`
	AutoCommit 	bool	`toml:"auto_commit"`
}

// APIInfo API config
type APIInfo struct {
	URL 		string	`toml:"url"`
	Method		string	`toml:"method"`
}



