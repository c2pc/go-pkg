create table if not exists auth_tasks
(
    id         serial
        primary key,
    user_id    integer not null references auth_users on update cascade on delete cascade,
    name       varchar(256) not null,
    type       varchar(256) not null,
    status     varchar(256) not null,
    output     bytea,
    input      bytea,
    created_at timestamp default now(),
    updated_at timestamp default now()
);