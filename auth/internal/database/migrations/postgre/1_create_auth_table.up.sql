create table if not exists auth_permissions
(
    id   serial
        primary key,
    name varchar(256)
        unique
);

create table if not exists auth_roles
(
    id   serial primary key,
    name varchar(256) unique
);

create table if not exists auth_role_permissions
(
    role_id       integer not null references auth_roles on update cascade on delete cascade,
    permission_id integer not null references auth_permissions on update cascade on delete cascade,
    read          boolean default false,
    write         boolean default false,
    exec          boolean default false,
    unique (role_id, permission_id)
);

create table if not exists auth_users
(
    id          serial primary key,
    login       varchar(256) not null UNIQUE,
    password    varchar(256) not null,
    email       varchar(256),
    phone       varchar(256),
    first_name  varchar(256) not null,
    second_name varchar(256),
    last_name   varchar(256),
    blocked     boolean default false
);

create table if not exists auth_user_roles
(
    user_id integer not null references auth_users on update cascade on delete cascade,
    role_id integer not null
        references auth_roles on update cascade on delete cascade,
    unique (user_id, role_id)
);

create table if not exists auth_tokens
(
    id         serial
        primary key,
    user_id    integer      not null references auth_users on update cascade on delete cascade,
    device_id  integer      not null,
    token      varchar(256) not null,
    updates_at timestamp    not null default now(),
    expires_at timestamp    not null default now(),
    logged_at  timestamp    not null default now(),
    unique (user_id, device_id)
);

create table if not exists auth_settings
(
    user_id   integer not null references auth_users on update cascade on delete cascade,
    device_id integer not null,
    settings  text,
    unique (user_id, device_id)
);

create table if not exists auth_tasks
(
    id         serial
        primary key,
    user_id    integer      not null
        references auth_users
            on update cascade on delete cascade,
    name       varchar(256) not null,
    type       varchar(256) not null,
    status     varchar(256) not null,
    output     bytea,
    input      bytea,
    created_at timestamp default now(),
    updated_at timestamp default now()
);