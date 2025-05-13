INSERT INTO users (id, name, email, roles, password_hash, department, enabled, created_at, updated_at) VALUES
	('5cf37266-3473-4006-984f-9325122678b7', 'Admin', 'admin@example.com', 'admin', '$2a$10$1ggfMVZV6Js0ybvJufLRUOWHS5f6KneuP0XwwHpJ8L8ipdry9f2/a', NULL, true, NOW(), NOW()),
	('45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'User', 'user@example.com', 'user', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NOW(), NOW())
ON DUPLICATE KEY UPDATE
     name = VALUES(name),
     email = VALUES(email),
     `roles` = VALUES(`roles`),
     password_hash = VALUES(password_hash),
     department = VALUES(department),
     enabled = VALUES(enabled),
     updated_at = VALUES(updated_at),
     created_at = VALUES(created_at);
