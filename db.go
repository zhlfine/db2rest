package main

import (
	"log"
	"database/sql"
	_ "github.com/lib/pq"
)

var pg *sql.DB

// InitDB init database
func InitDB() (err error) {
	pg, err = sql.Open("postgres", conf.DB.URL)
	if err == nil {
		pg.SetMaxIdleConns(conf.DB.MaxPoolSize)
		pg.SetMaxOpenConns(conf.DB.MaxPoolSize)
		pg.set
		err = pg.Ping()
	}
	return 
}

func CloseDB() (err error) {
	if pg != nil {
		pg.Close();
	}
}

func do(){
	db, err := sql.Open("postgres", conf.DB.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	orgID := "160090"
	rows, err := db.Query("select org_id, name from cloud_subscribers where org_id = $1 limit 3", orgID)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	var id, name string

	for rows.Next() {
		err := rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(id, name)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

}