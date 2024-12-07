CREATE TABLE auth_analytics (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED,
    operation_id VARCHAR(100) NOT NULl,
    method VARCHAR(10) NOT NULL,
    path TEXT NOT NULL,
    status_code INT NOT NULL,
    client_ip VARCHAR(45) NOT NULL,
    request_body BLOB,
    response_body BLOB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    first_name  varchar(256)         not null,
    second_name varchar(256)         null,
    last_name   varchar(256)         null,

    FOREIGN KEY (user_id) REFERENCES auth_users (id) ON DELETE SET NULl ON UPDATE CASCADE
);