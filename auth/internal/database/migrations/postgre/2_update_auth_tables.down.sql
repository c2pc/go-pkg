DELETE FROM auth_tokens WHERE LENGTH(token) > 256;

ALTER TABLE auth_tokens
    ALTER COLUMN token TYPE VARCHAR(256),
    DROP COLUMN IF EXISTS domain;