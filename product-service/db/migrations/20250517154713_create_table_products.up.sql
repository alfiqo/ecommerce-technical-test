-- Create products table
CREATE TABLE products (
    uuid            CHAR(36) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    base_price      DECIMAL(15, 2) NOT NULL,
    sku             VARCHAR(50) UNIQUE,
    barcode         VARCHAR(50) UNIQUE,
    weight          DECIMAL(10, 3),
    dimensions      VARCHAR(100),
    brand           VARCHAR(100),
    manufacturer    VARCHAR(100),
    category        VARCHAR(100),
    tags            VARCHAR(255),
    status          VARCHAR(50) NOT NULL DEFAULT 'active',
    image_urls      TEXT,
    thumbnail_url   VARCHAR(255),
    meta_title      VARCHAR(255),
    meta_description TEXT,
    meta_keywords   VARCHAR(255),
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (uuid)
) ENGINE = InnoDB;

-- Create product variants table
CREATE TABLE product_variants (
    uuid           CHAR(36) NOT NULL,
    product_uuid   CHAR(36) NOT NULL,
    sku            VARCHAR(50) UNIQUE,
    name           VARCHAR(255) NOT NULL,
    attributes     TEXT,
    price_diff     DECIMAL(15, 2) DEFAULT 0,
    thumbnail_url  VARCHAR(255),
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (uuid),
    FOREIGN KEY (product_uuid) REFERENCES products(uuid) ON DELETE CASCADE
) ENGINE = InnoDB;

-- Create product categories table
CREATE TABLE product_categories (
    uuid           CHAR(36) NOT NULL,
    name           VARCHAR(100) NOT NULL UNIQUE,
    description    TEXT,
    parent_uuid    CHAR(36),
    level          INT NOT NULL DEFAULT 0,
    image_url      VARCHAR(255),
    status         VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (uuid),
    FOREIGN KEY (parent_uuid) REFERENCES product_categories(uuid) ON DELETE SET NULL
) ENGINE = InnoDB;

-- Create product-category mappings table (many-to-many)
CREATE TABLE product_category_mappings (
    uuid           CHAR(36) NOT NULL,
    product_uuid   CHAR(36) NOT NULL,
    category_uuid  CHAR(36) NOT NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (uuid),
    UNIQUE KEY (product_uuid, category_uuid),
    FOREIGN KEY (product_uuid) REFERENCES products(uuid) ON DELETE CASCADE,
    FOREIGN KEY (category_uuid) REFERENCES product_categories(uuid) ON DELETE CASCADE
) ENGINE = InnoDB;