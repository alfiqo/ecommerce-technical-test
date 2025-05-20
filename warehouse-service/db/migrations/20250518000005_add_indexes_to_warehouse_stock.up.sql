-- Add indexes for better query performance on warehouse_stock table
ALTER TABLE warehouse_stock ADD INDEX idx_warehouse_product_combined (warehouse_id, product_id);
ALTER TABLE warehouse_stock ADD INDEX idx_product_id (product_id);