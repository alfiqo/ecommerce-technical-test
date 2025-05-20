-- Remove indexes added for warehouse_stock table
ALTER TABLE warehouse_stock DROP INDEX idx_warehouse_product_combined;
ALTER TABLE warehouse_stock DROP INDEX idx_product_id;