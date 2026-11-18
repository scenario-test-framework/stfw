#!/usr/bin/env bash
# scan-images.sh は Docker イメージを Trivy で検査し、合否の基準 (AS-BUILT §10.4) で判定する。
#
#   使い方: TRIVY_IMAGE=<image@digest> scan-images.sh <runtime-image> <full-image>
#
# 合否の基準:
#   - 自分たちで直せるもの (OS パッケージ・stfw 本体・同梱の compare-files / logfilter):
#     修正版があれば深刻度に関係なく失敗 (作り直し・版上げですぐ直せるため)
#   - 同梱の k6 (他社製): 修正版のある HIGH/CRITICAL のみ失敗
#     (部品の修正版が出ても k6 の新しい版が出るまで直せないため。中以下は定期チェックで報告)
#   - 修正版の無いものは直しようがないため合否に含めない (--ignore-unfixed)
#
# 終了コード: 0 = 合格 / 1 = 不合格 (または Trivy の実行失敗) / 2 = 使い方の誤り
set -u

: "${TRIVY_IMAGE:?TRIVY_IMAGE is required}"
if [ "$#" -ne 2 ]; then
  echo "usage: scan-images.sh <runtime-image> <full-image>" >&2
  exit 2
fi
runtime="$1"
full="$2"
# 脆弱性データベースを実行内の各検査で共有する (取得は最初の 1 回だけ。取得の制限による失敗を減らす)
cache="${TRIVY_CACHE_DIR:-${RUNNER_TEMP:-/tmp}/trivy-cache}"
mkdir -p "$cache"

trivy() {
  docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v "$cache:/root/.cache/" "$TRIVY_IMAGE" image \
    --scanners vuln --pkg-types os,library --ignore-unfixed \
    --exit-code 1 --no-progress "$@"
}

rc=0
for img in "$runtime" "$full"; do
  echo "=== ${img}: fixable findings at any severity (excluding bundled k6) ==="
  trivy --skip-dirs opt/stfw/bundled/k6 "$img" || rc=1
done
# k6 を含む全体を HIGH/CRITICAL で検査する (k6 以外は上の検査に含まれる)
echo "=== ${full}: fixable HIGH/CRITICAL findings (including bundled k6) ==="
trivy --severity HIGH,CRITICAL "$full" || rc=1

if [ "$rc" -ne 0 ]; then
  echo "::error::fixable vulnerabilities found (see the tables above)"
fi
exit "$rc"
