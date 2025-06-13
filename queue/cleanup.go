package queue

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func StartReservationCleaner(db *sqlx.DB, rdb *redis.Client) {
	ctx := context.Background()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cleanupExpiredReservations(ctx, db, rdb)
		}
	}
}

func cleanupExpiredReservations(ctx context.Context, db *sqlx.DB, rdb *redis.Client) {

	var cursor uint64
	var keys []string
	// using SCAN instead of KEYS for production safety
	for {
		batch, nextCursor, err := rdb.Scan(ctx, cursor, "reservation:*", 100).Result()
		if err != nil {
			log.Println("Redis scan error:", err)
			return
		}
		keys = append(keys, batch...)
		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	for _, key := range keys {
		ttl, err := rdb.TTL(ctx, key).Result()
		if err != nil {
			log.Println("TTL fetch error:", err)
			continue
		}

		if ttl > 0 {
			continue
		}
		if ttl == -1 {
			log.Printf("Reservation %s has no TTL. Cleaning up anyway.", key)
		}

		if !strings.HasPrefix(key, "reservation:") {
			log.Println("Skipping unrelated key:", key)
			continue
		}
		
		token := strings.TrimPrefix(key, "reservation:")
		mappingKey := fmt.Sprintf("token-ticket-type:%s", token)
		ticketTypeIDStr, err := rdb.Get(ctx, mappingKey).Result()
		if err == redis.Nil {
			log.Println("No ticketTypeID for expired token:", token)
			continue
		} else if err != nil {
			log.Println("Redis error on mapping get:", err)
			continue
		}

		ticketTypeID, err := strconv.Atoi(ticketTypeIDStr)
		if err != nil {
			log.Println("Invalid ticketTypeID in mapping:", ticketTypeIDStr)
			continue
		} 
		_, err = db.Exec("UPDATE ticket_types SET quantity_available = quantity_available + 1 WHERE id = $1", ticketTypeID)
		if err != nil {
			log.Println("DB stock refill error:", err)
			continue
		}

		log.Printf("Stock restored for ticket_type_id=%d after token %s expired\n", ticketTypeID, token)

		rdb.Del(ctx, key)
		rdb.Del(ctx, mappingKey)
	}
}
