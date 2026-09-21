INSERT INTO orders (id, status_code, customer_note, shipped_at, created_at, updated_at) VALUES
  (1, 1, NULL, NULL, NOW(), NOW()),
  (2, 2, '', NOW(), NOW(), NOW()),
  (3, 3, 'customer requested cancellation', NULL, NOW(), NOW()),
  (4, 9, NULL, NULL, NOW(), NOW()),
  (5, NULL, 'legacy data', NULL, NOW(), NOW());
