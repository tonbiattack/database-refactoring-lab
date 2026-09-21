CREATE TABLE order_statuses (
  code VARCHAR(32) NOT NULL PRIMARY KEY,
  name VARCHAR(64) NOT NULL UNIQUE
);

INSERT INTO order_statuses (code, name) VALUES
  ('accepted', '受付'),
  ('shipped', '発送済み'),
  ('canceled', 'キャンセル');
