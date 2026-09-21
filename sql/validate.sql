SELECT
  id,
  status_code,
  status,
  CASE
    WHEN status_code = 1 AND status = 'accepted' THEN 'Migrated'
    WHEN status_code = 2 AND status = 'shipped' THEN 'Migrated'
    WHEN status_code = 3 AND status = 'canceled' THEN 'Migrated'
    WHEN (status_code = 9 OR status_code IS NULL) AND status IS NULL THEN 'Unknown'
    ELSE 'Inconsistent'
  END AS classification
FROM orders
ORDER BY id;
