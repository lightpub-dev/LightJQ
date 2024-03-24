# Job queue 要件

## 用語
- JQ master: pusher からの job の enqueue, worker への job の分配, Web UI のための API 提供 を行うサーバー 
- Job: worker によって扱われる処理単位
- Worker: JQ master によって指示された　job を実行するクライアント。
- Pusher: JQ master に job を enqueue するクライアント。Worker と Pusher 両方になることも可能。
- constraint: 特定のJobMatcher に対してrate limit などの制限をかけられる
- JobMatcher: job.name や job.parameter などの条件を組み合わせて job を特定するマッチャー
- priority: job が持つ整数。この値が小さな job ほど優先される。32-bit signed integer.

## Job について
Job が持つ属性は次の通り
- name: job の種類。一意性がある。
- argument: job 実行時のパラメータ。1つの MessagePack オブジェクトで表現される。
- priority: 32-bit signed integer. 値が小さいほど優先度が高い。デフォルトは 0
- max_retry: リトライ回数の上限。0 ならリトライしない。リトライは exponential backoff。
- keep_result: boolean. job の実行結果を pusher が要求する予定なら true。つまり、true ならば JQ master は worker から返された結果オブジェクトを保持しておくべき。falseならば、jobの結果は利用されないので即時破棄してよい。
- timeout: タイムアウト。単位は秒。これを越した場合、Job は失敗したとみなしてよい。デフォルトは 30 秒。

これらの属性を持つ MessagePack オブジェクトが pusher から JQ master に送信され、enqueue される。

Job の実行結果として、次のMessagePackオブジェクトが worker から JQ master に送信される。属性は次の通り
成功時:
- type: "success"
- finished_at: job 終了時刻。ISO 8601
- result: 任意のMessagePack Value (array や number でも可)

失敗時:
- type: "failure"
- reason: "other" | "timeout"
- finished_at: job 終了時刻。ISO 8601
- should_retry: boolean これ以上リトライすべきか
- error: 任意の MessagePack Value
- message: Web UI などでの表示が想定されたエラーメッセージ

## 機能
- pusher は Job を JQ master に enqueue できる。JQ master は Pusher に新たに割り当てられた Job ID を返す。
- JQ master は enqueue された job を worker に割り振る
- JQ master は worker から job 実行結果を受け取る
- JQ master は keep_result=true に対して pusher からの結果要求に応える。一度要求が来たら結果は破棄してよい。また、keep_result=trueにも関わらず pusher からの結果要求が一定時間以上ない場合、結果を破棄してよい。破棄後に要求が来た場合は null を返すこと。

## Job の選択方法
次に 実行すべき (= worker に分配すべき) chosen_job は次のアルゴリズムで決める。
- S := (一度も実行されていない job) | (実行したが失敗した job) とする
- min_priority := min([j.priority | j ∈ S])
- chosen_job := choice([j | j ∈ S, j.priority == min_priority])
ただし、 choice は実装依存の関数 (ランダム選択、enqueueが最も古いもの などでよい)。

上記の仕様だと、priority が高いものは永遠に実行されない starvation が発生しうるが問題ない。

## JQ master との通信

### JQ master と pusher
JQ master が HTTP API を提供する。エンコーディングは MessagePack で行う。
仕様は OpenAPI により定義される。

### JQ master と worker
worker が HTTP API を提供する。エンコーディングは MessagePack で行う。
仕様は OpenAPI により定義される。

## 実装詳細
- MySQL はできれば使わず、Redis だけで完結したい
