package models

import "time"

type Event struct {
	ID          int       `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	EventDate   time.Time `db:"event_date"`
	EventTime   time.Time `db:"event_time"`
	Venue       string    `db:"venue"`
	CreatedAt   time.Time `db:"created_at"`
}

type TicketType struct {
	ID                int       `db:"id"`
	EventID           int       `db:"event_id"`
	Name              string    `db:"name"`
	Price             float64   `db:"price"`
	QuantityAvailable int       `db:"quantity_available"`
	CreatedAt         time.Time `db:"created_at"`
}

type QueueEntry struct {
	ID           int       `db:"id"`
	TicketTypeID int       `db:"ticket_type_id"`
	UserID       string    `db:"user_id"`
	Position     int       `db:"position"`
	JoinedAt     time.Time `db:"joined_at"`
}

type Reservation struct {
	ID           int       `db:"id"`
	TicketTypeID int       `db:"ticket_type_id"`
	UserID       string    `db:"user_id"`
	Status       string    `db:"status"`
	ReservedAt   time.Time `db:"reserved_at"`
	ExpiresAt    time.Time `db:"expires_at"`
}
