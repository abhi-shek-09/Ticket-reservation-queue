# Queue-Based Ticket Booking System

A scalable, real-time ticket reservation system built with Go, Redis, and PostgreSQL. Users join a queue for tickets, and background processors manage reservations, stock, and cleanup.

---

## Features

- **Queue-Based Booking**
  - Users join a Redis-backed queue (`queue:<ticket_type_id>`) to request tickets.
  - Background processor dequeues users and attempts to reserve tickets.
- **Reservation System**
  - Dequeued users are given a temporary reservation stored as a Redis hash (`reservation:<token>`).
  - Reservations auto-expire after 2 minutes.
- **Stock Management**
  - Stock is managed in PostgreSQL via the `ticket_types` table.
  - On successful reservation, DB stock is decreased by 1.
  - Reservation status and mapping are tracked via Redis with TTL.
- **Auto Expiry & Stock Refill**
  - Cleanup goroutine checks for expired reservations every 30 seconds.
  - If a reservation expires (TTL = 0), DB stock is refilled by +1.
  - Expired Redis keys (`reservation:*` and mapping keys) are deleted.
- **HTTP API Endpoints**
  - **POST /queue/join**: Join a Redis queue for a ticket.
  - **GET /reservation/status**: Check if a reservation is confirmed.
  - **POST /reservation/confirm**: Confirm reservation and finalize ticket.
  - **GET /events**: List all events (with ticket types).
- **Reservation Lifecycle**
  - User joins queue → stored in Redis list `queue:<ticket_type_id>`
  - Processor dequeues token, checks DB stock.
  - If stock available:
    - Reserve ticket.
    - Decrease stock in DB.
    - Store Redis key `reservation:<token>` with TTL = 2 min.
    - Map token to ticket type (`token-ticket-type:<token>`).
  - If reservation not confirmed within 2 min:
    - Cleaner detects TTL expiry.
    - Restores DB stock.
    - Deletes Redis keys.
  - If reservation confirmed:
    - `/reservation/confirm` finalizes it (optional: mark status in DB).

---

## Technologies

- **Go**: HTTP server, Redis & SQL handling, background goroutines.
- **Redis**: Queue (LPUSH/LPOP), reservation hashes, TTL, cleanup.
- **PostgreSQL**: Ticket types table for real stock control.
- **Gorilla Mux**: HTTP routing.
- **sqlx**: Simplified DB access.
- **go-redis**: Redis client.

---

## How to Run

1. **Start Redis and PostgreSQL**
   - Ensure both services are running and properly configured.
2. **Load DB Schema**
   - Load your database schema (includes `ticket_types`, `events`, etc.).
3. **Run the Server**

```go run main.go```

- The queue processor and cleaner run automatically as background goroutines.

---

## API Reference

| Method | Route                   | Description                           |
|--------|-------------------------|---------------------------------------|
| POST   | /queue/join             | Join a Redis queue for a ticket       |
| GET    | /reservation/status     | Check if a reservation is confirmed   |
| POST   | /reservation/confirm    | Confirm reservation and finalize ticket|
| GET    | /events                 | List all events (with ticket types)   |
