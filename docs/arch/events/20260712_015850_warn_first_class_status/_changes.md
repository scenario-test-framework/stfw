# 変更サマリ

- event_id: 20260712_015850_warn_first_class_status
- trigger_event: usdm:20260712_015850_warn_first_class_status

## 追加

- なし

## 変更

- レイヤールール LR-004（不正な状態遷移は error 返却）: 状態遷移の列挙へ Warn を追加
  （階層実行ステータス Started → Success / Warn / Error、ステップ実行ステータス
  Pending → Success / Warn / Error / Blocked）。Warn 一級ステータス導入
  （REQ-023 / SPEC-023-01/02）に伴う追従。外部レビュー指摘による追補。
- 集約仮説 AG-001（Run）: 不変条件を改訂。停止条件を「Error（終了コード 0・3 以外、
  または compare の on_mismatch=error での比較不一致）」に限定し、終了コード 3 = Warn の
  記録・続行を追加。終了状態の列挙を Success / Warn / Error に更新。
- ティアポリシー SP-004（実行順序保証・エラー時停止）: 停止条件を Error に限定し、
  Warn の記録・続行と Error > Warn > Success 集約を追記。
- 横断ポリシー CTP-005（再実行は新 run_id の別実行）: 終了状態の列挙を
  Success / Warn / Error に更新。

## 削除

- なし
