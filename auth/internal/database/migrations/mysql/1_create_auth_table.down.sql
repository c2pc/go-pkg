drop table if exists auth_settings;
drop table if exists auth_tokens;
drop table if exists auth_user_roles;
drop table if exists auth_users;
drop table if exists auth_role_permissions;
drop table if exists auth_roles;
drop table if exists auth_permissions;

DELETE FROM abonent_tokens
WHERE CHAR_LENGTH(refresh_token) > 256;

ALTER TABLE abonent_tokens
DROP COLUMN IF EXISTS domain,
    MODIFY COLUMN token VARCHAR(256),
    MODIFY COLUMN refresh_token VARCHAR(256);

drop table if exists auth_filters;