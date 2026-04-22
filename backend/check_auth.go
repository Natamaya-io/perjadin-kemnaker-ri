package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	dbUrl := "postgres://perjadin_user:Yuxt2MaJxNhpfvJzCyNEbVTG4JgtpwXH@localhost:5433/perjadin_db?sslmode=disable"
	conn, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		log.Fatal("Ping error: ", err)
	}

	rows, err := conn.Query("SELECT email, role, password FROM users LIMIT 5")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var email, role, password string
		if err := rows.Scan(&email, &role, &password); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("User: %s, Role: %s\n", email, role)
	}
}
