-- +goose Up
CREATE TABLE chirps (
    id uuid PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
	body TEXT NOT NULL,
	user_id uuid REFERENCES users(id) ON DELETE CASCADE
);


-- +goose Down
DROP TABLE chirps;
