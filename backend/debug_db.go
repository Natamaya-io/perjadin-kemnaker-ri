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
	users, err := q.GetUsers(context.Background())
	if err != nil {
		fmt.Println("Error getting users:", err)
		return
	}

	for _, u := range users {
		if u.Role == "protokol" {
			fmt.Printf("User: %s | Email: %s | ID: %s | Role: %s\n", u.Name, u.Email, u.ID, u.Role)
		}
	}

	records, err := q.GetTravelRecords(context.Background(), "")
	if err != nil {
		fmt.Println("Error getting records:", err)
		return
	}
	for _, r := range records {
		fmt.Printf("Record SPD: %s | EmpID: %s | CreatorID: %s\n", r.SpdNumber.String, r.EmployeeID, r.CreatorID)
	}
}
