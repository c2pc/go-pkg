create table if not exists news
(
    id   serial
        primary key,
    user_id integer not null references auth_users on update cascade on delete cascade,
    title varchar(256) unique NOT NULL ,
    content text NULL
);