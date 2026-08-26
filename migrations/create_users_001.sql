CREATE TABLE users (
	id SERIAL PRIMARY KEY,
	username TEXT UNIQUE,
	password_hash TEXT,
	created_at TIMESTAMP DEFAULT NOW()
)