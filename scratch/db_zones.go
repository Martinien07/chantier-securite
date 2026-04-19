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
	rows, err := db.Query("SELECT id, name, polygon FROM zones")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name, polygon string
		rows.Scan(&id, &name, &polygon)
		fmt.Printf("Zone ID: %d, Name: %s, Polygon: %s\n", id, name, polygon)
	}
}
