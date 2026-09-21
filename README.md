# database-refactoring-lab

数値の status_code を文字列の status へ安全に移す、DBリファクタリングの学習用プロジェクトです。

最終スキーマではなく、Expand、Dual Write、Backfill、Validate、Switch、Contractの順で互換性を維持する過程を扱います。

## 前提

- Docker Compose
- Go 1.18以上

初期化します。

~~~bash
cp .env.example .env
docker compose up -d mysql
docker compose run --rm flyway -target=2 migrate
docker compose run --rm flyway -target=4 migrate
~~~

Goコマンドの前に、環境変数を読み込みます。

~~~bash
set -a
. ./.env
set +a
~~~

最初からやり直す場合は、DBボリュームを削除します。

~~~bash
docker compose down -v
~~~

## Scenario A: Legacy DBを調査する

V002まで適用した状態では、正常な1、2、3に加えて、未定義の9とNULLがあります。

~~~bash
docker compose exec -T mysql mysql -uapp -papp refactoring_lab < sql/investigate.sql
~~~

customer_noteのNULL、空文字、通常値も別々に集計されます。

## Scenario B: Expandする

V003とV004でorder_statusesとNULL許容のstatusを追加します。

~~~bash
docker compose run --rm flyway -target=4 migrate
~~~

この時点ではNOT NULLも外部キーもありません。旧アプリケーションはstatus_codeだけを使い続けられます。

## Scenario C: Dual Writeする

APIを起動します。

~~~bash
WRITE_MODE=dual READ_MODE=fallback go run ./cmd/api
~~~

別の端末から注文を更新します。

~~~bash
curl -X PUT http://localhost:8080/orders/1/status \
  -H 'Content-Type: application/json' \
  -d '{"status":"shipped"}'
~~~

dualではstatus_codeとstatusが同じUPDATEで更新されます。legacyへ切り替えると、status_codeだけを更新します。

~~~bash
WRITE_MODE=legacy READ_MODE=fallback go run ./cmd/api
~~~

## Scenario D: Backfillする

既知のコードだけを2件ずつBackfillします。

~~~bash
go run ./cmd/backfill --batch-size=2
go run ./cmd/backfill --batch-size=2
~~~

2回目はUpdated: 0となります。Backfillはstatus IS NULLだけを対象とするため、再実行可能です。9とNULLは変換されません。

ID範囲を絞る場合は次を使います。

~~~bash
go run ./cmd/backfill --batch-size=100 --from-id=1 --to-id=1000
~~~

## Scenario E: Validateする

~~~bash
go run ./cmd/validate
~~~

初期サンプルをBackfillした結果は次です。

~~~text
Total:          5
Migrated:       3
Inconsistent:   0
Unknown:        2
~~~

不整合を意図的に作ると、Validateが検出します。

~~~bash
docker compose exec -T mysql mysql -uapp -papp refactoring_lab \
  -e "UPDATE orders SET status='shipped' WHERE status_code=1"
go run ./cmd/validate
~~~

id=1が対象ならInconsistentは1になります。検証後は、初期化してScenario Aからやり直してください。

Backfill中にlegacy書き込みを続けると、同じ問題が起きます。Backfill後にWRITE_MODE=legacyでstatus_codeだけを変更すると、新旧の値がずれます。すべての書き込み元をWRITE_MODE=dualへ切り替えてからBackfillすることで、新しい不整合を防げます。

## Scenario F: Switchする

Backfill後にInconsistentが0で、既知の既存データにstatusがない行がなく、すべての書き込み元がdualになっていることを確認してから切り替えます。

~~~bash
READ_MODE=new WRITE_MODE=dual go run ./cmd/api
~~~

newモードはstatus_codeへフォールバックしません。

## Scenario G: Contractする

Unknownが残る初期データのまま、V005を適用するとNOT NULL制約の追加が失敗します。

~~~bash
docker compose run --rm flyway -target=5 migrate
~~~

Unknownを自動変換してはいけません。次は学習用に、調査の結果としてcanceledを選んだ例です。本番では業務上の意味を確認してから決めます。

~~~bash
docker compose exec -T mysql mysql -uapp -papp refactoring_lab \
  -e "UPDATE orders SET status='canceled' WHERE id IN (4, 5)"
docker compose run --rm flyway repair
docker compose run --rm flyway -target=5 migrate
docker compose run --rm flyway -target=6 migrate
~~~

V006後はstatus_codeがないため、legacy書き込みとfallback読み取りは使えません。
