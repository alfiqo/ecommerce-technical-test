CREATE TABLE warehouse_stock (
    id                 INT UNSIGNED NOT NULL AUTO_INCREMENT,
    warehouse_id       INT UNSIGNED NOT NULL,
    product_id         INT UNSIGNED NOT NULL,
    quantity           INT NOT NULL DEFAULT 0,
    reserved_quantity  INT NOT NULL DEFAULT 0,
    -- Note: available_quantity is not stored but computed on the fly
    updated_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY unique_warehouse_product (warehouse_id, product_id),
    CONSTRAINT fk_warehouse_stock_warehouse_id FOREIGN KEY (warehouse_id) REFERENCES warehouses (id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB;