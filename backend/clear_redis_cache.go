package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/pkg/cache"
)

func main() {
	// Set host to localhost for external connection
	os.Setenv("REDIS_HOST", "localhost")
	
	cfg := config.LoadConfig()
	rdb, err := cache.NewRedisClient(cfg)
	if err != nil {
		fmt.Println("Err connecting to redis:", err)
		return
	}
	
	err = rdb.FlushAll(context.Background()).Err()
	if err != nil {
		fmt.Println("Err flushing redis:", err)
	} else {
		fmt.Println("Redis cache cleared successfully")
	}
}