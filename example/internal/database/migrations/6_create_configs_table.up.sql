create table if not exists auth_configs
(
    key        VARCHAR(256) PRIMARY KEY NOT NULL UNIQUE,
    value      bytea                    NOT NULL,
    updated_at timestamptz DEFAULT now()
);