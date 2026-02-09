-- +goose Up
ALTER TABLE users
ADD is_chirpy_red BOOLEAN DEFAULT false;

-- +goose Down
ALTER TABLE users
DROP CULOMN is_chirpy_red;