CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    Username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(Username);
CREATE INDEX idx_users_created_at ON users(created_at);