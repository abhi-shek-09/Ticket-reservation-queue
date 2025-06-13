package routes

import (
	"queue-system/handlers"
	"github.com/redis/go-redis/v9"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

func SetupRouter(db *sqlx.DB, rdb *redis.Client) *mux.Router {
	router := mux.NewRouter()
	
	router.HandleFunc("/events", handlers.GetAllEventsHandler(db)).Methods("GET")
	router.HandleFunc("/queue/join", handlers.JoinQueueHandler(db, rdb)).Methods("POST")
	router.HandleFunc("/reservation/status", handlers.CheckReservationStatusHandler(rdb)).Methods("GET")
	router.HandleFunc("/reservation/confirm", handlers.ConfirmReservationHandler(db, rdb)).Methods("POST")

	return router
}
