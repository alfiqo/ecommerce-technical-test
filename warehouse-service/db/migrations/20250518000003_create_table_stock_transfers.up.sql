CREATE TABLE stock_transfers (
    id                    INT UNSIGNED NOT NULL AUTO_INCREMENT,
    source_warehouse_id   INT UNSIGNED NOT NULL,
    target_warehouse_id   INT UNSIGNED NOT NULL,
    product_id            INT UNSIGNED NOT NULL,
    quantity              INT NOT NULL,
    status                ENUM('pending', 'completed', 'failed') NOT NULL DEFAULT 'pending',
    transfer_reference    VARCHAR(50) NOT NULL,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_transfer_reference (transfer_reference),
    CONSTRAINT fk_stock_transfers_source FOREIGN KEY (source_warehouse_id) REFERENCES warehouses (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_stock_transfers_target FOREIGN KEY (target_warehouse_id) REFERENCES warehouses (id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB;