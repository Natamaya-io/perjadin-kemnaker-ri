package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/db"
	"github.com/kemnaker/perjadin-backend/internal/domain/dalkot"
	"github.com/kemnaker/perjadin-backend/internal/domain/user"
)

func main() {
	connStr := "postgres://perjadin_stg_user:StgDB_xR98f2mQpW@perjadin-stg-db:5432/perjadin_stg_db?sslmode=disable"
	dbConn, err := db.NewPostgresDB(connStr)
	if err != nil {
		fmt.Println("DB ERROR:", err)
		os.Exit(1)
	}
	q := db.New(dbConn)
	repo := dalkot.NewRepository(q)
	userRepo := user.NewRepository(q)
	svc := dalkot.NewService(repo, userRepo)

	id, _ := uuid.Parse("a5050aef-cd02-45f3-9f38-9c1d00f09dc1")
	record, err := svc.GetRecordByID(id)
	if err != nil {
		fmt.Println("GetRecord ERROR:", err)
		os.Exit(1)
	}

	html, err := dalkot.ExportDprHTMLMock(record)
	if err != nil {
		fmt.Println("HTML ERROR:", err)
		os.Exit(1)
	}
	fmt.Println("SUCCESS, len:", len(html))
}
