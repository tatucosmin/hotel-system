-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(320) UNIQUE NOT NULL,
    hashed_password VARCHAR(120) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE refresh_token (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    hashed_token VARCHAR(500) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 day',
    PRIMARY KEY (user_id, hashed_token)
);

CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_number VARCHAR(10) UNIQUE NOT NULL,
    type        TEXT CHECK (type IN ('single', 'double', 'suite', 'deluxe')) NOT NULL,
    price       NUMERIC(10,2) NOT NULL CHECK (price >= 0),
    status      TEXT CHECK (status IN ('available', 'booked', 'maintenance')) DEFAULT 'available',
    description TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_rooms_status ON rooms(status);

CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    room_id      UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    check_in     DATE NOT NULL,
    check_out    DATE NOT NULL CHECK (check_out > check_in),
    status       TEXT CHECK (status IN ('pending', 'confirmed', 'checked-in', 'checked-out', 'cancelled')) DEFAULT 'pending',
    total_price  NUMERIC(10,2) NOT NULL CHECK (total_price >= 0),
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_bookings_user ON bookings(user_id);
CREATE INDEX idx_bookings_status ON bookings(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users, refresh_token, rooms, bookings;
-- +goose StatementEnd
