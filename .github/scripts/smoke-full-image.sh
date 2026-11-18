#!/usr/bin/env bash
# smoke-full-image.sh は stfw:full の中で、組込みプラグインが呼ぶコマンドが揃って動くかを確かめる
# (build.yml。ネットワークなしで実行する)。ベースの更新や Chrome の版上げで部品が欠けたら失敗させる。
#
#   使い方: smoke-full-image.sh <image>   (リポジトリ直下で実行)
#
# 終了コード: 0 = 合格 / 1 = 不合格 / 2 = 使い方の誤り
set -u

if [ "$#" -ne 1 ]; then
  echo "usage: smoke-full-image.sh <image>" >&2
  exit 2
fi
img="$1"
here="$(cd "$(dirname "$0")" && pwd)"
rc=0

echo "=== commands used by the built-in plugins"
# 各コマンドの版の表示が成功すること (共有ライブラリの欠けも分かる)。sed は GNU 版であること
docker run --rm --network none --entrypoint bash "$img" -c '
  set -e
  mysql --version
  psql --version
  redis-cli --version
  sshpass -V | head -1
  ssh -V
  sed --version | head -1 | grep -q "GNU sed"
  echo "GNU coreutils/sed: ok"
' || rc=1

echo "=== k6 browser mode with the bundled Chrome"
docker run --rm --network none -v "${here}/smoke-browser.js:/smoke/browser.js:ro" --entrypoint bash "$img" -c '
  set -e
  k6="$(ls /opt/stfw/bundled/k6/*/k6 | head -1)"
  "${K6_BROWSER_EXECUTABLE_PATH}" --version
  "${k6}" run --quiet /smoke/browser.js
' || rc=1

if [ "$rc" -ne 0 ]; then
  echo "::error::stfw:full smoke test failed (see the output above)"
fi
exit "$rc"
