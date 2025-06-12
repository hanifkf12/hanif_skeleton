-- +goose Up
-- +goose StatementBegin
CREATE TABLE users(
    id int auto_increment primary key,
    name varchar(100) not null,
    email varchar(100) not null unique,
    username varchar(50) not null unique,
    created_at timestamp default now(),
    updated_at timestamp default now() on update now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE if exists users;
-- +goose StatementEnd 