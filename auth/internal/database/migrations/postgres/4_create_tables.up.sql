-- =====================================================
-- auth_users
-- =====================================================
ALTER TABLE IF EXISTS auth_users
    ALTER COLUMN password DROP NOT NULL;

-- =====================================================
-- auth_analytics_admins
-- =====================================================
ALTER TABLE IF EXISTS auth_analytics_admins
    ADD COLUMN IF NOT EXISTS action     TEXT,
    ADD COLUMN IF NOT EXISTS error      TEXT,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT now();

DO
$$
    BEGIN
        IF EXISTS (SELECT 1
                   FROM information_schema.columns
                   WHERE table_name = 'auth_analytics_admins'
                     AND column_name = 'first_name') THEN
            ALTER TABLE IF EXISTS auth_analytics_admins
                RENAME COLUMN first_name TO "name";
        END IF;
    END;
$$;

DO
$$
    BEGIN
        IF EXISTS (SELECT 1
                   FROM information_schema.columns
                   WHERE table_name = 'auth_analytics_admins'
                     AND column_name = 'name') THEN
            ALTER TABLE IF EXISTS auth_analytics_admins
                ALTER COLUMN name TYPE VARCHAR(768) USING name::VARCHAR(768),
                ALTER COLUMN name DROP NOT NULL;
        END IF;
    END;
$$;

ALTER TABLE IF EXISTS auth_analytics_admins
    DROP COLUMN IF EXISTS second_name,
    DROP COLUMN IF EXISTS last_name,
    DROP COLUMN IF EXISTS duration;
