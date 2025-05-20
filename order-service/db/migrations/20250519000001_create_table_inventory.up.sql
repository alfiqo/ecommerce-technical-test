CREATE TABLE IF NOT EXISTS inventory (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    product_id INT UNSIGNED NOT NULL,
    warehouse_id INT UNSIGNED NOT NULL,
    quantity INT NOT NULL,
    reserved_quantity INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Each product can only have one inventory entry per warehouse
    UNIQUE KEY idx_product_warehouse (product_id, warehouse_id),
    
    -- Indexes for faster queries
    INDEX idx_product_id (product_id),
    INDEX idx_warehouse_id (warehouse_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add comment explaining the purpose of this table
ALTER TABLE inventory COMMENT 'Stores product inventory with strict locking for stock management';