-- Version: 1.01
-- Description: Create table users
CREATE TABLE users
(
    id            CHAR(36)     NOT NULL,
    name          VARCHAR(250) NOT NULL,
    email         VARCHAR(150) NOT NULL,
    roles SET ('user', 'admin') NOT NULL,
    password_hash VARCHAR(150) NOT NULL,
    department    VARCHAR(200) NULL,
    enabled       BOOLEAN      NOT NULL,
    refresh_token VARCHAR(255) NULL,
    updated_at    TIMESTAMP(6) NOT NULL,
    created_at    TIMESTAMP(6) NOT NULL,

    PRIMARY KEY (id)
) ENGINE = InnoDB
  DEFAULT CHARSET = latin1
  COLLATE = latin1_general_ci;

-- Version: 1.02
-- Description: Create table products
CREATE TABLE products
(
    id         CHAR(36)       NOT NULL,
    name       VARCHAR(250)   NOT NULL,
    price      NUMERIC(10, 2) NOT NULL,
    updated_at TIMESTAMP(6)   NOT NULL,
    created_at TIMESTAMP(6)   NOT NULL,

    PRIMARY KEY (id)
) ENGINE = InnoDB
  DEFAULT CHARSET = latin1
  COLLATE = latin1_general_ci;

-- Version: 1.03
-- Description: Create table products
CREATE TABLE sales
(
    id         CHAR(36)     NOT NULL,
    user_id    CHAR(36)     NOT NULL,
    discount   NUMERIC(10, 2) NULL,
    amount     NUMERIC(10, 2) NULL,
    updated_at TIMESTAMP(6) NOT NULL,
    created_at TIMESTAMP(6) NOT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = latin1
  COLLATE = latin1_general_ci;

-- Version: 1.04
-- Description: Create table products
CREATE TABLE sale_items
(
    sale_id    CHAR(36)       NOT NULL,
    product_id CHAR(36)       NOT NULL,
    quantity   INT(3)   NOT NULL,
    discount   NUMERIC(10, 2) NULL,
    amount     NUMERIC(10, 2) NOT NULL,
    updated_at TIMESTAMP(6)   NOT NULL,
    created_at TIMESTAMP(6)   NOT NULL,

    PRIMARY KEY (sale_id, product_id),
    FOREIGN KEY (sale_id) REFERENCES sales (id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = latin1
  COLLATE = latin1_general_ci;

-- Version: 1.05
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