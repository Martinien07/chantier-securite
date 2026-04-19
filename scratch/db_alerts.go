package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "admin:admin@tcp(localhost:3306)/make_hse")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query("SELECT id, alert_level, status FROM alerts ORDER BY id DESC LIMIT 5")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var level, status sql.NullString
		rows.Scan(&id, &level, &status)
		fmt.Printf("Alert %d: level=%s, status=%s\n", id, level.String, status.String)
	}
}
