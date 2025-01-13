create table if not exists auth_profiles
(
    id      serial primary key,
    age     integer      null,
    height    integer      null,
    address varchar(256) not null,
    user_id integer      not null UNIQUE references auth_users on update cascade on delete cascade
);

