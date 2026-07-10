#!/usr/bin/env bash
# daily-balance example を end-to-end で実行する。
#   ./run.sh          依存起動 → secret 準備 → plugin install → run → レポート配信
#   ./run.sh --down   後片付け (コンテナ・ボリューム削除)
set -euo pipefail
cd "$(dirname "$0")"

if [ "${1:-}" = "--down" ]; then
  docker compose down -v
  exit 0
fi

run() { docker compose run --rm -T stfw "$@"; }

# plugin install は既に install 済みだと exit 3 (Warn) を返す。それは正常扱いにする。
ensure_plugin() { run plugin install "$1" || [ "$?" -eq 3 ]; }

echo "==> 依存サービス (postgres + トイ API = テスト対象、Jaeger = トレース送信先) を起動"
docker compose up -d --build postgres api jaeger

echo "==> secret を準備 (age 鍵 + DB パスワード)"
if [ ! -f stfw/config/encrypt/key.txt ]; then
  run secret keygen
fi
# DB ホスト postgres へ user/password を登録 (user は stfw.yml の stfw.db.user と一致)。
# (再実行でも冪等になるよう --force で上書き)
run secret set postgres appuser apppass --force

echo "==> プラグインの外部バイナリ (k6 / compare-files) を install"
ensure_plugin invokeRest
ensure_plugin compare

echo "==> シナリオ実行 (daily-balance)"
run run daily-balance

echo "==> HTML レポート配信 (nginx) を起動"
docker compose up -d nginx
echo
echo "    レポート: http://localhost:${STFW_REPORT_PORT:-8088}"
echo "    トレース: http://localhost:${STFW_JAEGER_PORT:-16686}  (Jaeger UI。service=stfw を選択)"
echo "    後片付け: ./run.sh --down"
