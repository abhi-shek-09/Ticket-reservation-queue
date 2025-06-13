package queue

import (
	"context"
	"fmt"
	"log"
	"time"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// This function is like a ticket clerk who wakes up every fixed time interval, checks who is in line (Redis queue), and starts processing them one by one for each ticket type
//  Every 1 minute...                                                                       
//  Loop over all `ticketTypeIDs`                                                           
//  For each one: launch a goroutine to process its queue                                   
//  Each goroutine pops 1 user, checks DB, reserves ticket, and stores reservation in Redis 

func StartQueueProcessor(DB *sqlx.DB, RDB *redis.Client, ticketTypeIDs []int) {
	ctx := context.Background() // Used for managing timeout, cancellation, etc

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop() //  ticker is stopped if the processor is ever shut down (cleanup)

	for{ 
		//  infinite loop that uses a select statement to wait for a signal every 10s
		select {
		case <-ticker.C: // When the ticker fires, it enters the block
			for _, ticketTypeID := range ticketTypeIDs {
				go processQueueForTicketType(ctx, DB, RDB, ticketTypeID) // concurrent and isolated
			}
		}
	}
} // handles multiple ticket types, Non-blocking , and checks the queue only periodically

func processQueueForTicketType(ctx context.Context, DB *sqlx.DB, RDB *redis.Client, ticketTypeID int) {
	queueKey := fmt.Sprintf("queue:%d", ticketTypeID)

	token, err := RDB.LPop(ctx, queueKey).Result()
	if err == redis.Nil {
		return
	} else if err != nil {
		log.Println("Redis LPOP error:", err)
		return
	}

	var available int
	err = DB.QueryRow("SELECT quantity_available FROM ticket_types WHERE id = $1", ticketTypeID).Scan(&available)
	if err != nil {
		log.Println("DB error:", err)
		return
	}

	if available <= 0 {
		log.Printf("No stock for ticket type %d\n", ticketTypeID)
		return
	}

	_, err = DB.Exec("UPDATE ticket_types SET quantity_available = quantity_available - 1 WHERE id = $1", ticketTypeID)
	if err != nil {
		log.Println("DB update error:", err)
		return
	}

	reservationKey := fmt.Sprintf("reservation:%s", token)
	mappingKey := fmt.Sprintf("token-ticket-type:%s", token)
	reservationTTL := 2 * time.Minute
	reservationData := map[string]any{
		"ticket_type_id": ticketTypeID,
		"timestamp":      time.Now().Unix(),
	}

	pipe := RDB.TxPipeline()
	pipe.HSet(ctx, reservationKey, reservationData)
	pipe.Expire(ctx, reservationKey, reservationTTL)
	pipe.Set(ctx, mappingKey, ticketTypeID, reservationTTL)
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Println("Failed to atomically set reservation and expiry:", err)
		return
	}

	log.Printf("Reserved for %s (ticket_type_id=%d)\n", token, ticketTypeID)
}
