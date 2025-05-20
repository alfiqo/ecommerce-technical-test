CREATE TABLE orders (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id         CHAR(36) NOT NULL,
    status          ENUM('pending', 'paid', 'cancelled', 'completed') NOT NULL DEFAULT 'pending',
    total_amount    DECIMAL(10, 2) NOT NULL,
    shipping_address TEXT NOT NULL,
    payment_method   VARCHAR(50) NOT NULL,
    payment_deadline TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status)
) ENGINE = InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;