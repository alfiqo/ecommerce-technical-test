CREATE TABLE IF NOT EXISTS stock_movements (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    warehouse_id INT UNSIGNED NOT NULL,
    product_id INT UNSIGNED NOT NULL,
    product_sku VARCHAR(100) NOT NULL,
    movement_type ENUM('stock_in', 'stock_out', 'transfer_in', 'transfer_out') NOT NULL,
    quantity INT NOT NULL,
    reference_type VARCHAR(50),
    reference_id VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_warehouse_product (warehouse_id, product_id),
    INDEX idx_product_sku (product_sku),
    INDEX idx_reference (reference_type, reference_id),
    CONSTRAINT fk_stock_movements_warehouse FOREIGN KEY (warehouse_id) REFERENCES warehouses (id) ON DELETE CASCADE
) ENGINE = InnoDB;