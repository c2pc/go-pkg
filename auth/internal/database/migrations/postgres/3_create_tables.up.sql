-- =====================================================
-- auth_permissions
-- =====================================================
CREATE TABLE IF NOT EXISTS auth_permissions
(
    id   SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL UNIQUE
);

-- =====================================================
-- auth_roles
-- =====================================================
CREATE TABLE IF NOT EXISTS auth_roles
(
    id   SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL UNIQUE
);

-- =====================================================
-- auth_role_permissions
-- =====================================================
CREATE TABLE IF NOT EXISTS auth_role_permissions
(
    role_id       INTEGER NOT NULL REFERENCES auth_roles (id) ON UPDATE CASCADE ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES auth_permissions (id) ON UPDATE CASCADE ON DELETE CASCADE,
    read          BOOLEAN DEFAULT FALSE,
    write         BOOLEAN DEFAULT FALSE,
    exec          BOOLEAN DEFAULT FALSE,
    UNIQUE (role_id, permission_id)
);

-- =====================================================
-- auth_users
-- =====================================================
CREATE TABLE IF NOT EXISTS auth_users
(
    id          BIGSERIAL PRIMARY KEY,
    login       VARCHAR(256) NOT NULL UNIQUE,
    password    VARCHAR(256),
    email       VARCHAR(256),
    phone       VARCHAR(256),
    first_name  VARCHAR(256) NOT NULL,
    second_name VARCHAR(256),
    last_name   VARCHAR(256),
    blocked     BOOLEAN DEFAULT FALSE
);

ALTER TABLE IF EXISTS auth_users
    ALTER COLUMN password DROP NOT NULL,
    ADD COLUMN IF NOT EXISTS is_domain BOOLEAN DEFAULT FALSE;

-- =====================================================
-- auth_user_roles
-- =====================================================
CREATE TABLE IF NOT EXISTS auth_user_roles
(
    user_id BIGINT NOT NULL REFERENCES auth_users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES auth_roles ON UPDATE CASCADE ON DELETE CASCADE,
    UNIQUE (user_id, role_id)
);

-- =====================================================
-- auth_tokens
-- =====================================================
CREATE TABLE IF NOT EXISTS auth_tokens
(
    id        BIGSERIAL PRIMARY KEY,
    user_id   BIGINT  NOT NULL REFERENCES auth_users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    device_id INTEGER NOT NULL,
    token     TEXT    NOT NULL,
    UNIQUE (user_id, device_id)
);

ALTER TABLE IF EXISTS auth_tokens
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    ADD COLUMN IF NOT EXISTS logged_at  TIMESTAMPTZ DEFAULT now() NOT NULL,
    ADD COLUMN IF NOT EXISTS provider   varchar(256);

-- =====================================================
-- auth_settings
-- =====================================================
CREATE TABLE IF NOT EXISTS auth_settings
(
    user_id   BIGINT  NOT NULL REFERENCES auth_users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    device_id INTEGER NOT NULL,
    UNIQUE (user_id, device_id)
);

ALTER TABLE IF EXISTS auth_settings
    ADD COLUMN IF NOT EXISTS settings BYTEA;

-- =====================================================
-- auth_analytics_admins
-- =====================================================
ALTER TABLE IF EXISTS auth_analytics
    RENAME TO auth_analytics_admins;

CREATE TABLE IF NOT EXISTS auth_analytics_admins
(
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT,
    login         VARCHAR(256),
    method        VARCHAR(10)  NOT NULL,
    operation_id  VARCHAR(100) NOT NULL,
    path          TEXT         NOT NULL,
    status_code   INTEGER      NOT NULL,
    client_ip     VARCHAR(45)  NOT NULL,
    request_body  BYTEA,
    response_body BYTEA,
    name          VARCHAR(768)
);

-- =====================================================
-- auth_filters
-- =====================================================
CREATE TABLE IF NOT EXISTS auth_filters
(
    id        BIGSERIAL PRIMARY KEY,
    user_id   BIGINT       NOT NULL REFERENCES auth_users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    device_id INTEGER      NOT NULL,
    endpoint  VARCHAR(256) NOT NULL,
    name      VARCHAR(256) NOT NULL,
    value     BYTEA        NOT NULL
);

-- =====================================================
-- auth_tasks
-- =====================================================
CREATE TABLE IF NOT EXISTS auth_tasks
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT       NOT NULL REFERENCES auth_users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    name       VARCHAR(256) NOT NULL,
    "type"     VARCHAR(256) NOT NULL,
    status     VARCHAR(256) NOT NULL,
    output     BYTEA,
    input      BYTEA,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- =====================================================
-- auth_configs
-- =====================================================
CREATE TABLE IF NOT EXISTS auth_configs
(
    key   VARCHAR(256) PRIMARY KEY,
    value BYTEA NOT NULL
);

ALTER TABLE IF EXISTS auth_configs
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT now();

-- =====================================================
-- auth_config_files
-- =====================================================
CREATE TABLE IF NOT EXISTS auth_config_files
(
    key        VARCHAR(256) PRIMARY KEY,
    config_key VARCHAR(256) NOT NULL REFERENCES auth_configs (key) ON UPDATE CASCADE ON DELETE CASCADE,
    file_name  TEXT         NOT NULL,
    value      BYTEA        NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- =====================================================
-- auth_users_blocked
-- =====================================================
CREATE TABLE IF NOT EXISTS auth_users_blocked
(
    user_id    BIGINT      NOT NULL UNIQUE REFERENCES auth_users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    blocked_at TIMESTAMPTZ NOT NULL
);

