package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/db"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.LoadConfig()

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	if os.Getenv("DATABASE_URL") != "" {
		dsn = os.Getenv("DATABASE_URL")
	}

	database, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer database.Close()

	if err := database.Ping(); err != nil {
		log.Fatalf("Ping failed: %v", err)
	}
	
	q := db.New(database)
	
	records, err := q.GetTravelRecords(context.Background(), "")
	if err != nil {
		fmt.Printf("GetTravelRecords error: %v\n", err)
		return
	}
	fmt.Printf("Total Travel Records: %d\n", len(records))
	for _, rec := range records {
	    if rec.SpdNumber.String == "ID-SPJ-055" {
		    fmt.Printf("Record %s found. ID: %s\n", rec.SpdNumber.String, rec.ID)
		    
		    cost, err := q.GetTravelCostByRecordID(context.Background(), rec.ID)
		    if err != nil {
		        fmt.Printf("  -> Cost: Error fetching (%v)\n", err)
		        
		        // Also query from raw database 
		        row := database.QueryRow("SELECT ticket_go, daily_allowance_rate, hotel_rate, local_transport FROM travel_costs WHERE travel_record_id = $1", rec.ID)
		        var tGo, dRate, hRate, lTrans sql.NullFloat64
		        ierr := row.Scan(&tGo, &dRate, &hRate, &lTrans)
		        if ierr != nil {
		            fmt.Printf("  -> RAW Cost Error: %v\n", ierr)
		        } else {
		            fmt.Printf("  -> RAW Cost: tGo=%v dRate=%v hRate=%v lTrans=%v\n", tGo.Float64, dRate.Float64, hRate.Float64, lTrans.Float64)
		        }
		    } else {
		        fmt.Printf("  -> DB Cost: ticket_go=%v, daily_rate=%v, hotel_rate=%v, local_trans=%v\n", cost.TicketGo.Float64, cost.DailyAllowanceRate.Float64, cost.HotelRate.Float64, cost.LocalTransport.Float64)
		    }
		}
	}
}
