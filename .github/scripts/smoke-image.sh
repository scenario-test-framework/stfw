#!/usr/bin/env bash
# shellcheck disable=SC2016 # イメージ内で実行するスクリプトは単一引用符で渡す (展開はコンテナの中で行う)
# smoke-image.sh は、イメージの中に stfw と組込みプラグインが呼ぶコマンドが揃い、起動できるかを確かめる
# (build.yml と scheduled-scan.yml。ネットワークなしで実行する)。Wolfi はパッケージが常に最新へ更新されるため、
# ベースの更新や Chrome の版上げで部品が欠けたら失敗させる (AS-BUILT §10.2 / §10.4)。
#
#   使い方: smoke-image.sh runtime <image>   通常版の確認
#           smoke-image.sh full <image>      通常版の確認 + full 版の部品 (DB クライアント・Chrome など)
#   (リポジトリ直下で実行)
#
# 終了コード: 0 = 合格 / 1 = 不合格 / 2 = 使い方の誤り
set -u

if [ "$#" -ne 2 ] || { [ "$1" != runtime ] && [ "$1" != full ]; }; then
  echo "usage: smoke-image.sh <runtime|full> <image>" >&2
  exit 2
fi
kind="$1"
img="$2"
here="$(cd "$(dirname "$0")" && pwd)"
rc=0

# in_image はイメージ内で bash スクリプトを実行する (失敗した行で止め、パイプの途中の失敗も拾う)
in_image() {
  docker run --rm --network none --entrypoint bash "$@"
}

echo "=== ${img}: commands used by stfw and the plugins (runtime)"
in_image "$img" -c '
  set -euo pipefail
  # stfw 本体と、プラグイン・stfw ssh trust・stfw plugin install が呼ぶコマンド
  for c in stfw bash curl ssh scp ssh-keygen ssh-keyscan tar gzip find sed grep awk sort mktemp uname diff; do
    command -v "$c" >/dev/null || { echo "missing command: $c" >&2; exit 1; }
  done
  stfw --version
  ssh -V
  # GNU 版の基本コマンド (wolfi-base の既定は BusyBox。Debian の頃と同じオプションを使えるようにしている)
  for c in sed grep find tar date awk diff; do
    "$c" --version 2>&1 | head -1 | grep -q GNU || { echo "not GNU: $c" >&2; exit 1; }
  done
  # TZ 指定が効くこと (tzdata が無いと黙って UTC になる)
  test "$(TZ=Asia/Tokyo date +%Z)" = JST || { echo "TZ=Asia/Tokyo is not applied (tzdata missing?)" >&2; exit 1; }
  echo "runtime commands: ok"
' || rc=1

if [ "$kind" = full ]; then
  echo "=== ${img}: commands used by the built-in plugins (full)"
  in_image "$img" -c '
    set -euo pipefail
    for c in mysql mysqldump psql redis-cli sshpass fc-list; do
      command -v "$c" >/dev/null || { echo "missing command: $c" >&2; exit 1; }
    done
    mysql --version
    psql --version
    redis-cli --version
    sshpass -V | head -1
    # MySQL 8 以降の認証方式の部品 (/etc/my.cnf の plugin-dir が指す先)
    test -f /usr/lib/mariadb/plugin/caching_sha2_password.so || { echo "missing caching_sha2_password.so" >&2; exit 1; }
    # invokeWeb の日本語表示に使うフォント
    fc-list :lang=ja family | grep -q . || { echo "no Japanese font visible to fontconfig" >&2; exit 1; }
    # compare / compareRegex が使う同梱の compare-files (compare_files と compare_regex は同じ版を共有する)
    for c in compare_files compare_regex; do
      b="$(ls /opt/stfw/bundled/compare-files/*/"$c" | head -1)"
      "$b" -v
    done
    echo "full commands: ok"
  ' || rc=1

  echo "=== ${img}: k6 browser mode with the bundled Chrome"
  in_image -v "${here}/smoke-browser.js:/smoke/browser.js:ro" "$img" -c '
    set -euo pipefail
    k6="$(ls /opt/stfw/bundled/k6/*/k6 | head -1)"
    "${K6_BROWSER_EXECUTABLE_PATH}" --version
    "${k6}" run --quiet /smoke/browser.js
  ' || rc=1
fi

if [ "$rc" -ne 0 ]; then
  echo "::error::smoke test of ${img} (${kind}) failed (see the output above)"
fi
exit "$rc"
