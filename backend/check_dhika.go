package main

import (
	"context"
	"fmt"
	"database/sql"
	"github.com/kemnaker/perjadin-backend/internal/db"
	_ "github.com/lib/pq"
)

func main() {
	dbUrl := "postgres://perjadin_user:Yuxt2MaJxNhpfvJzCyNEbVTG4JgtpwXH@localhost:5433/perjadin_db?sslmode=disable"
	conn, err := sql.Open("postgres", dbUrl)
	if err != nil {
		fmt.Println("Error connecting to db:", err)
		return
	}
	defer conn.Close()

	q := db.New(conn)
	users, _ := q.GetUsers(context.Background())
	var dhikaID string
	for _, u := range users {
		if u.Email == "dhikanurkhaliffa" {
			dhikaID = u.ID.String()
			fmt.Printf("Dhika ID: %s\n", dhikaID)
		}
	}

	records, _ := q.GetTravelRecords(context.Background(), "")
	found := 0
	for _, r := range records {
		if r.EmployeeID.String() == dhikaID {
			fmt.Printf("FOUND record for Dhika: SPD=%s, ID=%s\n", r.SpdNumber.String, r.ID)
			found++
		}
	}
	fmt.Printf("Total records for Dhika: %d\n", found)
}
