package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/jmoiron/sqlx"
)

type QueueJoinRequest struct {
	UserID       string `json:"user_id"`
	TicketTypeID int    `json:"ticket_type_id"`
}

type QueueJoinResponse struct {
	QueueToken string `json:"queue_token"`
	Message    string `json:"message"`
}

func JoinQueueHandler(db *sqlx.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		var req QueueJoinRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid Request", http.StatusBadRequest)
			return
		}

		var count int // checking if a ticket type exists
		err := db.QueryRow("SELECT COUNT(*) FROM ticket_types WHERE id = $1", req.TicketTypeID).Scan(&count)
		if err != nil || count == 0 {
			http.Error(w, "Invalid ticket type", http.StatusBadRequest)
			return
		}

		// gen a token
		timestamp := time.Now().UnixNano()
		token := fmt.Sprintf("%s:%d", req.UserID, timestamp)

		// push to redis queue
		// gen a queue key thats diff for every type of ticket
		// this way a person paying a higher amount (considering that fewer ppl want it) can get it faster
		queueKey := fmt.Sprintf("queue:%d", req.TicketTypeID)
		if err := rdb.RPush(ctx, queueKey, token).Err(); err != nil {
			http.Error(w, "Failed to enqueue", http.StatusInternalServerError)
			return
		}

		resp := QueueJoinResponse{
			QueueToken: token,
			Message:    "You have joined the queue. Please wait for your turn.",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}