drop table if exists auth_settings;
drop table if exists auth_tokens;
drop table if exists auth_user_roles;
drop table if exists auth_users;
drop table if exists auth_role_permissions;
drop table if exists auth_roles;
drop table if exists auth_permissions;

DELETE FROM auth_tokens WHERE LENGTH(token) > 256;

ALTER TABLE auth_tokens
ALTER COLUMN token TYPE VARCHAR(256),
    DROP COLUMN IF EXISTS domain;

drop table if exists auth_filters;