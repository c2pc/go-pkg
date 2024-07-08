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
    blocked     boolean default false,
    settings    text
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
    user_id    integer      not null references auth_users on update cascade on delete cascade,
    device_id  integer      not null,
    token      varchar(256) not null,
    expires_at timestamp    not null,
    unique (user_id, device_id)
);

