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
	rows, err := db.Query("SELECT c.name, cc.pts_plan, cc.pts_image FROM camera_calibrations cc JOIN cameras c ON cc.camera_id = c.id")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name, plan, img string
		rows.Scan(&name, &plan, &img)
		fmt.Printf("Camera: %s\nPlan: %s\nImg: %s\n", name, plan, img)
	}
}
