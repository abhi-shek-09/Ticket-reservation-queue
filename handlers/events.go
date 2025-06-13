package handlers

import (
	"encoding/json"
	"net/http"
	"queue-system/models"
	"github.com/jmoiron/sqlx"
)

type EventWithTickets struct {
	models.Event
	TicketTypes []models.TicketType `json:"ticket_types"`
}


func GetAllEventsHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, description, event_date, event_time, venue, created_at FROM events ORDER BY event_date")
		if err != nil {
			http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
			return
		}
		defer rows.Close()


		var events []EventWithTickets
		for rows.Next(){
			var event models.Event
			if err := rows.Scan(&event.ID, &event.Name, &event.Description, &event.EventDate, &event.EventTime, &event.Venue, &event.CreatedAt); err != nil {
				http.Error(w, "Error scanning events", http.StatusInternalServerError)
				return
			}
			
			ticketRows, err := db.Query("SELECT id, event_id, name, price, quantity_available, created_at FROM ticket_types WHERE event_id = $1", event.ID)
			if err != nil {
				http.Error(w, "Error fetching tickets", http.StatusInternalServerError)
				return
			}

			var tickets []models.TicketType
			for ticketRows.Next() {
				var ticket models.TicketType
				err := ticketRows.Scan(&ticket.ID, &ticket.EventID, &ticket.Name, &ticket.Price, &ticket.QuantityAvailable, &ticket.CreatedAt)
				if err != nil {
					http.Error(w, "Error scanning ticket", http.StatusInternalServerError)
					return
				}
				tickets = append(tickets, ticket)
			}
			ticketRows.Close()

			events = append(events, EventWithTickets{Event: event, TicketTypes: tickets})
		}

		w.Header().Set("Content-type", "application/json")
		json.NewEncoder(w).Encode(events)
	}
}

