package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type ReservationStatusResponse struct {
	Status       string `json:"status"` // reserved, pending, expired
	TicketTypeID int    `json:"ticket_type_id,omitempty"`
	Timestamp    int64  `json:"timestamp,omitempty"`
}

type ConfirmReservationRequest struct {
	Token string `json:"token"`
}

// write the http handlerfunc inside our router function
//  Aspect                Explanation                                                                                                        
//  Dependency Injection  You inject `redisClient`, `db`, etc., instead of accessing them as globals. Much cleaner, testable, and decoupled. 
//  Context Handling      You control the root context (`ctx`) from the top level — useful for cancellation, timeouts, or tracing.           
//  Consistency           Follows the pattern of how middlewares are built in Go (`http.HandlerFunc` with closures).                        
//  Testability           Easier to write unit tests for `ConfirmReservationHandler(...)` since it's just returning a function.              


func CheckReservationStatusHandler(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If you're querying the DB or Redis without context, the operation might keep running wasting memory and resources.
		// But if you pass a context and the request is cancelled Then downstream Redis or DB calls can abort early
		// A context can be request-scoped(dies w the request), operation-scoped, manually cancellable
		// Avoid blocking forever if Redis is slow or down, Cancel queries if they are stuck or slow, Give them scoped lifetimes if needed, Always scoped to the request lifecycle
		
		ctx := context.Background()
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Missing token", http.StatusBadRequest)
			return
		}

		key := fmt.Sprintf("reservation:%s", token)
		data, err := rdb.HGetAll(ctx, key).Result()
		if err != nil {
			http.Error(w, "Redis error", http.StatusInternalServerError)
			return
		}
		ticketTypeIDStr := r.URL.Query().Get("ticket_type_id")

		if len(data) == 0 {
			// check if still in queue
			queueKey := fmt.Sprintf("queue:%s", ticketTypeIDStr)
			inQueue, _ := rdb.LRange(ctx, queueKey, 0, -1).Result() // get all tokens for a ticket type
			for _, t := range inQueue {
				if t == token {
					json.NewEncoder(w).Encode(ReservationStatusResponse{Status: "pending"})
					return
				}
			}
			// Not reserved, not in queue => expired
			json.NewEncoder(w).Encode(ReservationStatusResponse{Status: "expired"})
			return
		}

		ticketTypeID, _ := strconv.Atoi(data["ticket_type_id"])
		timestamp, _ := strconv.ParseInt(data["timestamp"], 10, 64)

		json.NewEncoder(w).Encode(ReservationStatusResponse{
			Status:       "reserved",
			TicketTypeID: ticketTypeID,
			Timestamp:    timestamp,
		})
	}
}

func ConfirmReservationHandler(db *sqlx.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		ctx := r.Context() // use the request's context

		var req ConfirmReservationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		reservationKey := fmt.Sprintf("reservation:%s", req.Token)
		resData, err := rdb.HGetAll(ctx, reservationKey).Result()
		if err != nil || len(resData) == 0{
			http.Error(w, "Reservation expired or not found", http.StatusNotFound)
			return
		}

		userID := strings.Split(req.Token, ":")[0]
		ticketTypeIDStr := resData["ticket_type_id"]
		ticketTypeID, _ := strconv.Atoi(ticketTypeIDStr)

		_, err = db.Exec("INSERT INTO ticket_bookings (user_id, ticket_type_id) VALUES ($1, $2)", userID, ticketTypeID)
		if err != nil {
			http.Error(w, "Failed to confirm booking", http.StatusInternalServerError)
			return
		}
		
		mappingKey := fmt.Sprintf("token-ticket-type:%s", req.Token)
		rdb.Del(ctx, reservationKey, mappingKey)

		json.NewEncoder(w).Encode(map[string]string {
			"message" : "Your booking has been confirmed!!",
		})
	}
	
}