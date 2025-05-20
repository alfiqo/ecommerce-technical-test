CREATE TABLE IF NOT EXISTS reservation_logs (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    warehouse_id INT UNSIGNED NOT NULL,
    product_id INT UNSIGNED NOT NULL,
    quantity INT NOT NULL,
    status ENUM('pending', 'committed', 'cancelled') NOT NULL DEFAULT 'pending',
    reference VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_warehouse_product (warehouse_id, product_id),
    INDEX idx_reference (reference),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE CASCADE
);