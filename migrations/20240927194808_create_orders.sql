-- +goose Up
-- +goose StatementBegin
create table if not exists orders (
    id bigint not null,
    client_id bigint not null,
    pickup_point_id bigint not null,
    status text not null,
    status_updated timestamptz not null,
    weight bigint not null default 0,
    cost bigint not null default 0,
    primary key (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- drop table if exists orders;
-- +goose StatementEnd
