create table if not exists auth_permissions
(
    id   bigint unsigned auto_increment
        primary key,
    name varchar(256) null,
    constraint name
        unique (name)
);

create table if not exists auth_roles
(
    id   bigint unsigned auto_increment
        primary key,
    name varchar(256) null,
    constraint name
        unique (name)
);

create table if not exists auth_role_permissions
(
    role_id       bigint unsigned      not null,
    permission_id bigint unsigned      not null,
    `read`        tinyint(1) default 0 null,
    `write`       tinyint(1) default 0 null,
    `exec`        tinyint(1) default 0 null,
    constraint role_id
        unique (role_id, permission_id),
    constraint auth_role_permissions_ibfk_1
        foreign key (role_id) references auth_roles (id)
            on update cascade on delete cascade,
    constraint auth_role_permissions_ibfk_2
        foreign key (permission_id) references auth_permissions (id)
            on update cascade on delete cascade
);

create index permission_id
    on auth_role_permissions (permission_id);

create table if not exists auth_users
(
    id          bigint unsigned auto_increment
        primary key,
    login       varchar(256)         not null,
    password    varchar(256)         not null,
    email       varchar(256)         null,
    phone       varchar(256)         null,
    first_name  varchar(256)         not null,
    second_name varchar(256)         null,
    last_name   varchar(256)         null,
    blocked     tinyint(1) default 0 null,
    settings    text                 null,
    constraint login
        unique (login)
);

create table if not exists auth_settings
(
    user_id   bigint unsigned not null,
    device_id int             not null,
    settings  text            null,
    constraint user_id
        unique (user_id, device_id),
    constraint auth_settings_ibfk_1
        foreign key (user_id) references auth_users (id)
            on update cascade on delete cascade
);

create table if not exists auth_tokens
(
    id         bigint unsigned auto_increment
        primary key,
    user_id    bigint unsigned                       not null,
    device_id  int                                   not null,
    token      varchar(256)                          not null,
    updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    expires_at timestamp default current_timestamp() not null on update current_timestamp(),
    logged_at  timestamp default current_timestamp() not null on update current_timestamp(),
    constraint user_id
        unique (user_id, device_id),
    constraint auth_tokens_ibfk_1
        foreign key (user_id) references auth_users (id)
            on update cascade on delete cascade
);

create table if not exists auth_user_roles
(
    user_id bigint unsigned not null,
    role_id bigint unsigned not null,
    constraint user_id
        unique (user_id, role_id),
    constraint auth_user_roles_ibfk_1
        foreign key (user_id) references auth_users (id)
            on update cascade on delete cascade,
    constraint auth_user_roles_ibfk_2
        foreign key (role_id) references auth_roles (id)
            on update cascade on delete cascade
);

create index role_id
    on auth_user_roles (role_id);

create table if not exists auth_tasks
(
    id         bigint unsigned auto_increment
        primary key,
    user_id    bigint unsigned                       not null,
    name       varchar(256)                          not null,
    type       varchar(256)                          not null,
    status     varchar(256)                          not null,
    output     longblob                              null,
    input      longblob                              null,
    created_at timestamp default current_timestamp() not null,
    updated_at timestamp default current_timestamp() not null,
    constraint auth_tasks_ibfk_1
        foreign key (user_id) references auth_users (id)
            on update cascade on delete cascade
);

ALTER TABLE abonent_tokens
    ADD COLUMN IF NOT EXISTS domain BOOLEAN NOT NULL DEFAULT FALSE,
    MODIFY COLUMN token TEXT,
    MODIFY COLUMN refresh_token TEXT;


create table if not exists auth_filters
(
    id        bigint unsigned auto_increment primary key,
    user_id   bigint unsigned not null,
    device_id int             not null,
    endpoint  varchar(256)    not null,
    name      varchar(256)    not null,
    value     longblob        not null,
    constraint auth_filters_user_ibfk_1
        foreign key (user_id) references auth_users (id)
            on update cascade on delete cascade
)