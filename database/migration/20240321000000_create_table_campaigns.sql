-- +goose Up
-- +goose StatementBegin
CREATE TABLE campaigns(
    id char(36) primary key,
    name varchar(100) not null,
    target_donation decimal(15,2) not null,
    end_date date not null,
    created_at timestamp default now(),
    updated_at timestamp default now() on update now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE if exists campaigns;
-- +goose StatementEnd