package main

import (
	"fmt"
	"log"
	"net/http"

	"queue-system/config"
	"queue-system/db"
	"queue-system/redis"
	"queue-system/routes"
	"queue-system/queue"
)

func main() {
	cfg := config.Load()

	dbConn := db.InitPostgres(cfg)
	redisClient := redis.InitRedis(cfg)

	defer dbConn.Close()
	defer redisClient.Close()
	go queue.StartQueueProcessor(dbConn, redisClient, []int{1, 2, 3})
	router := routes.SetupRouter(dbConn, redisClient)
	go queue.StartReservationCleaner(dbConn, redisClient)

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}