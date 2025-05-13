-- Version: 1.01
-- Description: Create table users
CREATE TABLE users
(
    user_id       CHAR(36)     NOT NULL,
    name          VARCHAR(250) NOT NULL,
    email         VARCHAR(150) NOT NULL,
    mobile        VARCHAR(150) NULL,
    profile_image TEXT NULL,
    roles SET ('user', 'staff', 'manager', 'support', 'admin') NOT NULL,
    password_hash VARCHAR(150) NOT NULL,
    department    VARCHAR(200) NULL,
    enabled       BOOLEAN      NOT NULL,
    refresh_token VARCHAR(255) NULL,
    updated_at    TIMESTAMP(6) NOT NULL,
    created_at    TIMESTAMP(6) NOT NULL,

    PRIMARY KEY (user_id)
) ENGINE = InnoDB
  DEFAULT CHARSET = latin1
  COLLATE = latin1_general_ci;

-- Version: 1.02
-- Description: Create table products
CREATE TABLE products
(
    product_id CHAR(36)       NOT NULL,
    user_id    CHAR(36)       NOT NULL,
    name       VARCHAR(250)   NOT NULL,
    cost       NUMERIC(10, 2) NOT NULL,
    quantity   INT            NOT NULL,
    updated_at TIMESTAMP(6)   NOT NULL,
    created_at TIMESTAMP(6)   NOT NULL,

    PRIMARY KEY (product_id),
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = latin1
  COLLATE = latin1_general_ci;

-- Version: 1.03
-- Description: Add products view.
CREATE OR REPLACE VIEW view_products AS
SELECT p.product_id,
       p.user_id,
       p.name,
       p.cost,
       p.quantity,
       p.created_at,
       p.updated_at,
       u.name AS user_name
FROM products AS p
         LEFT JOIN users AS u ON u.user_id = p.user_id

-- Version: 1.04
-- Description: Add password reset table
CREATE TABLE password_reset_tokens
(
    email     VARCHAR(150) NOT NULL,
    token     VARCHAR(255) NOT NULL,
    expiry_at TIMESTAMP(6) NOT NULL,
    KEY (email)
) ENGINE = InnoDB
  DEFAULT CHARSET = latin1
  COLLATE = latin1_general_ci;