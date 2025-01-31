DELETE FROM abonent_tokens
WHERE CHAR_LENGTH(refresh_token) > 256;

ALTER TABLE abonent_tokens
    DROP COLUMN IF EXISTS domain,
    MODIFY COLUMN refresh_token VARCHAR(256);