CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    event_date DATE NOT NULL,
    event_time TIME NOT NULL,
    venue TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ticket_types (
    id SERIAL PRIMARY KEY,
    event_id INTEGER REFERENCES events(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    quantity_available INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS queues (
    id SERIAL PRIMARY KEY,
    ticket_type_id INTEGER REFERENCES ticket_types(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    position INTEGER NOT NULL,
    joined_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ticket_bookings (
    id SERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    ticket_type_id INTEGER NOT NULL REFERENCES ticket_types(id),
    reserved_at TIMESTAMP NOT NULL DEFAULT NOW()
);
