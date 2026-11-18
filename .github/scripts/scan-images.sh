#!/usr/bin/env bash
# scan-images.sh は Docker イメージを Trivy で検査し、合否の基準 (AS-BUILT §10.4) で判定する。
#
#   使い方: TRIVY_IMAGE=<image@digest> scan-images.sh <image>...   (リポジトリ直下で実行)
#
# 合否の基準:
#   - 修正版がある脆弱性は、深刻度に関係なく失敗 (すべての部品で同じ)
#   - ただし除外リスト (.trivyignore.yaml) にある「対応版待ち」は合否から外す。
#     除外したものも結果に毎回表示する (--show-suppressed)
#   - 修正版の無いものは直しようがないため合否に含めない (--ignore-unfixed)
#   - 除外リスト自体が決まり (場所・期限・理由) を守っていなければ失敗 (check-trivyignore.sh)
#
# 終了コード: 0 = 合格 / 1 = 不合格 (または Trivy の実行失敗) / 2 = 使い方の誤り
set -u

if [ -z "${TRIVY_IMAGE:-}" ]; then
  echo "TRIVY_IMAGE is required (image@digest of aquasec/trivy)" >&2
  exit 2
fi
if [ "$#" -lt 1 ]; then
  echo "usage: scan-images.sh <image>..." >&2
  exit 2
fi
here="$(cd "$(dirname "$0")" && pwd)"
ignorefile="${TRIVY_IGNOREFILE:-.trivyignore.yaml}"
# 脆弱性データベースを実行内の各検査で共有する (取得は最初の 1 回だけ。取得の制限による失敗を減らす)
cache="${TRIVY_CACHE_DIR:-${RUNNER_TEMP:-/tmp}/trivy-cache}"
mkdir -p "$cache"

rc=0
echo "=== ${ignorefile}: checking the rules ==="
bash "${here}/check-trivyignore.sh" "$ignorefile" || rc=1

ignore_dir="$(cd "$(dirname "$ignorefile")" && pwd)"
ignore_name="$(basename "$ignorefile")"
for img in "$@"; do
  echo "=== ${img}: fixable findings at any severity (listed in ${ignore_name} are shown as suppressed) ==="
  docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v "$cache:/root/.cache/" \
    -v "${ignore_dir}:/trivyignore:ro" "$TRIVY_IMAGE" image \
    --scanners vuln --pkg-types os,library --ignore-unfixed \
    --ignorefile "/trivyignore/${ignore_name}" --show-suppressed \
    --exit-code 1 --no-progress "$img" || rc=1
done

if [ "$rc" -ne 0 ]; then
  echo "::error::image scan failed: fix the findings, or add a reasoned entry to ${ignore_name} (see the output above)"
fi
exit "$rc"
