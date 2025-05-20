-- Create shop_warehouses junction table for warehouse association
CREATE TABLE IF NOT EXISTS shop_warehouses (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    shop_id BIGINT UNSIGNED NOT NULL,
    warehouse_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_shop_warehouse (shop_id, warehouse_id),
    INDEX idx_warehouse_id (warehouse_id),
    CONSTRAINT fk_shop_warehouses_shop_id FOREIGN KEY (shop_id) REFERENCES shops (id) ON DELETE CASCADE
) ENGINE=InnoDB;