# database-refactoring-lab 仕様書

## 1. 目的

運用中のデータベースを安全にリファクタリングする流れを、ローカル環境で再現・検証できる学習用サンプルプロジェクトとする。

単に最終的なスキーマを示すのではなく、既存データと旧アプリケーション互換性を維持しながら、

1. Expand
2. Dual Write
3. Backfill
4. Validate
5. Switch
6. Contract

の順に段階的に移行するプロセスを体験できることを目的とする。

題材は、数値の `status_code` で注文状態を管理している既存システムを、意味の分かる文字列ステータスと参照テーブルへ移行するケースとする。

---

## 2. 想定技術スタック

- Go
- MySQL 8
- Flyway
- Docker Compose
- database/sql
- MySQL Driver for Go

ORMは使用せず、SQLを明示的に扱う。

---

## 3. 学習テーマ

このプロジェクトでは以下を確認できるようにする。

- 既存データの分布を調査してから変更する
- 古い構造をすぐ削除せず、新旧構造を一時的に共存させる
- Backfillより先にDual Writeを導入する理由
- Backfillを再実行可能にする方法
- 想定外データを無理に正常値へ変換しない方法
- 新旧データの不整合をSQLで検出する方法
- 制約を段階的に追加する方法
- 最後に旧カラムを削除するContractフェーズ
- Flywayでマイグレーション履歴を管理する方法

---

## 4. 初期状態

初期の `orders` テーブルは、意図的に改善余地のある構造とする。

```sql
CREATE TABLE orders (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  status_code INT NULL,
  customer_note VARCHAR(255) NULL,
  shipped_at DATETIME NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
```

### status_code

| 値 | 意味 |
|---|---|
| 1 | accepted |
| 2 | shipped |
| 3 | canceled |
| 9 | 未定義値 |
| NULL | 状態不明 |

`9` と `NULL` は、Backfill時に自動変換せず検出対象とする。

### customer_note

以下の3状態をサンプルデータに含める。

- NULL
- 空文字
- 通常の文字列

NULLと空文字が実際に異なる分布を持つことを確認するために使用する。

---

## 5. サンプルデータ

最低限、以下を含める。

```text
id=1 status_code=1 customer_note=NULL
id=2 status_code=2 customer_note=''
id=3 status_code=3 customer_note='customer requested cancellation'
id=4 status_code=9 customer_note=NULL
id=5 status_code=NULL customer_note='legacy data'
```

正常データだけでなく、意図的に未定義値を含める。

---

## 6. フェーズ構成

### Phase 1: Investigate

現在のデータ分布を確認する。

例:

```sql
SELECT
  status_code,
  COUNT(*) AS count
FROM orders
GROUP BY status_code
ORDER BY status_code;
```

NULLと空文字についても確認する。

```sql
SELECT
  CASE
    WHEN customer_note IS NULL THEN 'null'
    WHEN customer_note = '' THEN 'empty'
    ELSE 'has_value'
  END AS note_state,
  COUNT(*) AS count
FROM orders
GROUP BY note_state;
```

調査用SQLは `sql/investigate.sql` に配置する。

---

### Phase 2: Expand

既存構造を削除せず、新しい構造だけ追加する。

追加するテーブル:

```sql
CREATE TABLE order_statuses (
  code VARCHAR(32) NOT NULL PRIMARY KEY,
  name VARCHAR(64) NOT NULL UNIQUE
);
```

初期データ:

```text
accepted / 受付
shipped / 発送済み
canceled / キャンセル
```

`orders` に新しいカラムを追加する。

```sql
ALTER TABLE orders
  ADD COLUMN status VARCHAR(32) NULL;
```

この段階では以下を付与しない。

- NOT NULL
- FOREIGN KEY

旧アプリケーションが `status_code` のみを使用しても動作できる状態を維持する。

---

### Phase 3: Dual Write

Goアプリケーションから注文状態を更新する際、新旧両方のカラムを更新する。

例:

```sql
UPDATE orders
SET
  status_code = ?,
  status = ?,
  updated_at = NOW()
WHERE id = ?;
```

Go側では、ステータスの対応を1箇所に集約する。

例:

```go
type Status struct {
    LegacyCode int
    Code       string
}
```

最低限、以下を扱う。

- accepted
- shipped
- canceled

未定義値はDual Writeの対象外とする。

---

## 7. API

簡易HTTP APIを用意する。

### GET /orders/{id}

注文を1件取得する。

Switch前は以下の順で値を返す。

1. `status` が存在する場合は `status`
2. NULLの場合は `status_code` から変換

### PUT /orders/{id}/status

状態を更新する。

リクエスト例:

```json
{
  "status": "shipped"
}
```

Dual Write期間中は、新旧両方へ更新する。

### GET /health

動作確認用。

---

## 8. Backfill

既存の `status_code` から `status` へデータを移行する。

実行方法:

```bash
go run ./cmd/backfill
```

変換ルール:

```text
1 -> accepted
2 -> shipped
3 -> canceled
その他 -> 更新しない
```

### 要件

Backfillは以下を満たすこと。

- 再実行可能
- `status IS NULL` の行だけ対象
- 主キー範囲で分割可能
- 1回の処理件数を指定可能
- 未定義値は無理に変換しない

例:

```bash
go run ./cmd/backfill --batch-size=100
```

大量データを一度に更新しない設計とする。

---

## 9. Validate

Backfill後、新旧データの整合性を確認する。

実行方法:

```bash
go run ./cmd/validate
```

出力例:

```text
Total:          5
Migrated:       3
Inconsistent:   0
Unknown:        2
```

### 判定

正常:

```text
1 / accepted
2 / shipped
3 / canceled
```

不整合例:

```text
1 / shipped
2 / canceled
```

Unknown:

```text
9 / NULL
NULL / NULL
```

不整合検出SQLは `sql/validate.sql` に置く。

---

## 10. 意図的な不整合の再現

学習用として、不整合を作る方法もREADMEに記載する。

例:

```sql
UPDATE orders
SET status = 'shipped'
WHERE status_code = 1;
```

その状態で `cmd/validate` を実行すると、

```text
Inconsistent: 1
```

となることを確認する。

---

## 11. Backfill中の更新競合の再現

Dual Writeが必要な理由を確認するため、以下を再現できるようにする。

1. Legacy Writeモードでアプリを起動
2. Backfillを実行
3. Backfill中に注文を更新
4. `status_code` だけ更新される
5. 新旧カラムに不整合が発生する
6. Validateで検出する

その後Dual Writeモードに切り替え、同様の操作で不整合が発生しないことを確認する。

---

## 12. Write Mode

環境変数で書き込み方式を切り替え可能にする。

```env
WRITE_MODE=legacy
```

または

```env
WRITE_MODE=dual
```

### legacy

`status_code` のみ更新する。

### dual

`status_code` と `status` の両方を更新する。

この機能は移行途中の旧アプリケーションを再現するために使用する。

---

## 13. Read Mode

読み取り方式も環境変数で変更可能にする。

```env
READ_MODE=fallback
```

または

```env
READ_MODE=new
```

### fallback

`status` を優先し、NULLの場合は `status_code` から変換する。

### new

`status` のみ利用する。

Switchフェーズを再現するために使用する。

---

## 14. Switch

以下の状態を確認した後、Read Modeを `new` に切り替える。

- Backfill完了
- Inconsistent = 0
- 正常な既存データで `status IS NULL` が存在しない
- Dual Writeが有効

Switch後は旧 `status_code` を読み取りに使用しない。

---

## 15. Contract

最終フェーズで制約を追加する。

### NOT NULL

```sql
ALTER TABLE orders
  MODIFY COLUMN status VARCHAR(32) NOT NULL;
```

### FOREIGN KEY

```sql
ALTER TABLE orders
  ADD CONSTRAINT fk_orders_status
  FOREIGN KEY (status)
  REFERENCES order_statuses(code);
```

### 旧カラム削除

```sql
ALTER TABLE orders
  DROP COLUMN status_code;
```

Contractは、未定義データが解消されていることを確認した後に実行する。

初期サンプルでは `status_code=9` やNULLが存在するため、そのままではNOT NULL化できないことも確認できるようにする。

---

## 16. Flywayマイグレーション

想定構成:

```text
migrations/
├── V001__create_legacy_orders.sql
├── V002__insert_sample_data.sql
├── V003__create_order_statuses.sql
├── V004__add_status_column.sql
├── V005__add_status_constraints.sql
└── V006__drop_status_code.sql
```

BackfillはFlywayのDDLマイグレーションへ埋め込まず、Goコマンドとして実装する。

理由:

- 小分け実行しやすくする
- 再実行可能にする
- 大量更新をDDL適用と分離する
- 実行状況を観測しやすくする

---

## 17. ディレクトリ構成

```text
database-refactoring-lab/
├── README.md
├── SPEC.md
├── compose.yaml
├── go.mod
├── .env.example
├── migrations/
│   ├── V001__create_legacy_orders.sql
│   ├── V002__insert_sample_data.sql
│   ├── V003__create_order_statuses.sql
│   ├── V004__add_status_column.sql
│   ├── V005__add_status_constraints.sql
│   └── V006__drop_status_code.sql
├── cmd/
│   ├── api/
│   │   └── main.go
│   ├── backfill/
│   │   └── main.go
│   └── validate/
│       └── main.go
├── internal/
│   ├── db/
│   │   └── db.go
│   └── order/
│       ├── model.go
│       ├── repository.go
│       ├── service.go
│       └── status.go
└── sql/
    ├── investigate.sql
    └── validate.sql
```

---

## 18. Docker Compose

最低限以下のサービスを用意する。

- mysql
- flyway

MySQLは8系を使用する。

起動例:

```bash
docker compose up -d mysql
```

マイグレーション例:

```bash
docker compose run --rm flyway migrate
```

---

## 19. 環境変数

```env
DB_HOST=localhost
DB_PORT=3306
DB_NAME=refactoring_lab
DB_USER=app
DB_PASSWORD=app

WRITE_MODE=dual
READ_MODE=fallback
```

`.env.example` のみGit管理する。

---

## 20. READMEに記載するシナリオ

READMEでは、以下を順に実行できるようにする。

### Scenario A: Legacy DBの確認

- DB起動
- Legacy migration適用
- データ分布確認

### Scenario B: Expand

- 新テーブル、新カラム追加
- 旧アプリがそのまま動くことを確認

### Scenario C: Dual Write

- APIで更新
- 新旧両方に値が入ることを確認

### Scenario D: Backfill

- Backfill実行
- 正常な3件だけ移行されることを確認

### Scenario E: Validate

- 未定義値がUnknownとして残ることを確認
- 意図的に不整合を作って検出する

### Scenario F: Switch

- READ_MODE=newへ変更
- 新カラムのみで読み取る

### Scenario G: Contract

- 不正データを解消
- NOT NULL / FOREIGN KEY追加
- status_code削除

---

## 21. 非目標

このプロジェクトでは以下は扱わない。

- 大規模本番環境向けの完全なオンラインDDL検証
- レプリケーション環境
- CDC
- Debezium
- Kafka
- シャーディング
- 本格的なBlue/Green Deployment
- Kubernetes
- ORM比較

あくまで、DBリファクタリングの基本的な移行手順を理解するための最小構成とする。

---

## 22. 完了条件

以下を満たしたら初期版完成とする。

- Docker ComposeだけでMySQLを起動できる
- FlywayでLegacyスキーマを作成できる
- 意図的な未定義データを含むサンプルデータが存在する
- APIからLegacy Write / Dual Writeを切り替えられる
- Read Modeをfallback / newで切り替えられる
- Backfillを再実行可能
- Validateで不整合と未定義値を検出できる
- 制約追加前に失敗条件を確認できる
- 最後にstatus_codeを削除できる
- READMEの手順だけで一連の移行を再現できる

---

## 23. 設計方針

このプロジェクトでは、最終スキーマの美しさよりも、そこへ安全に到達する過程を重視する。

特に以下を明示的に再現する。

- Expand before Contract
- Dual Write before Backfill
- Backfill must be retryable
- Unknown values must remain observable
- Validate before Switch
- Contract only after compatibility is no longer required

DBリファクタリングを「DDLを書く作業」ではなく、「互換性を保った状態遷移の設計」として理解できるサンプルを目指す。
