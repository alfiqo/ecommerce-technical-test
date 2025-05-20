-- Drop tables in reverse order of creation to handle foreign key constraints
DROP TABLE IF EXISTS product_category_mappings;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS product_categories;
DROP TABLE IF EXISTS products;