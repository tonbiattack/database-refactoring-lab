ALTER TABLE orders
  MODIFY COLUMN status VARCHAR(32) NOT NULL;

ALTER TABLE orders
  ADD CONSTRAINT fk_orders_status
  FOREIGN KEY (status)
  REFERENCES order_statuses(code);
