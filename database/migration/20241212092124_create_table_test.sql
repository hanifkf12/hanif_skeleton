-- +goose Up
-- +goose StatementBegin
CREATE TABLE test(
                     id int auto_increment primary key ,
                     name varchar(100),
                     create_at timestamp default now()
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE if exists test;
-- +goose StatementEnd
