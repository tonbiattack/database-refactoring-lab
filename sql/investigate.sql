SELECT
  status_code,
  COUNT(*) AS count
FROM orders
GROUP BY status_code
ORDER BY status_code;

SELECT
  CASE
    WHEN customer_note IS NULL THEN 'null'
    WHEN customer_note = '' THEN 'empty'
    ELSE 'has_value'
  END AS note_state,
  COUNT(*) AS count
FROM orders
GROUP BY note_state
ORDER BY note_state;
