#!/usr/bin/env bash
# check-trivyignore-stale.sh は、除外リスト (.trivyignore.yaml) のうち、もうどのイメージでも
# 見つからなくなった行 (対応版を取り込んだ後の消し忘れ) を警告する (AS-BUILT §10.4 の定期チェック)。
#
#   使い方: TRIVY_IMAGE=<image@digest> check-trivyignore-stale.sh <image>...   (リポジトリ直下で実行)
#
# 警告の注記を出すだけで、常に終了コード 0 を返す (消し忘れは合否に影響しないため)。
# 検査に失敗したときは全行を消し忘れと誤報しないよう、警告して打ち切る (pipefail で docker の失敗も拾う)。
set -u -o pipefail

: "${TRIVY_IMAGE:?TRIVY_IMAGE is required}"
here="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=trivyignore-lib.sh
. "${here}/trivyignore-lib.sh"
ignorefile="${TRIVY_IGNOREFILE:-.trivyignore.yaml}"
for tool in yq jq; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo "::warning::stale check: ${tool} is required; skipped"
    exit 0
  fi
done
cache="${TRIVY_CACHE_DIR:-${RUNNER_TEMP:-/tmp}/trivy-cache}"
mkdir -p "$cache"

# 除外リストを使わずに検査し、見つかった「場所<TAB>ID」を集める
found=""
for img in "$@"; do
  if ! out="$(docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v "$cache:/root/.cache/" "$TRIVY_IMAGE" image \
    --scanners vuln --pkg-types os,library --ignore-unfixed --format json --no-progress --quiet "$img" \
    | jq -r '.Results[]? | .Target as $t | .Vulnerabilities[]? | "\($t)\t\(.VulnerabilityID)"')"; then
    echo "::warning::stale check: failed to scan ${img}; skipped"
    exit 0
  fi
  found+="${out}"$'\n'
done

count="$(yq '.vulnerabilities | length' "$ignorefile")"
stale=0
for ((i = 0; i < count; i++)); do
  id="$(yq ".vulnerabilities[$i].id" "$ignorefile")"
  npaths="$(yq ".vulnerabilities[$i].paths // [] | length" "$ignorefile")"
  hit=0
  while IFS=$'\t' read -r target fid; do
    [ "$fid" = "$id" ] || continue
    for ((j = 0; j < npaths; j++)); do
      pattern="$(yq ".vulnerabilities[$i].paths[$j]" "$ignorefile")"
      # Trivy と同じく * は 1 階層だけに一致させる (trivyignore-lib.sh)
      if ti_path_match "$pattern" "$target"; then
        hit=1
        break 2
      fi
    done
  done <<<"$found"
  if [ "$hit" -eq 0 ]; then
    echo "::warning file=${ignorefile}::${id} is no longer found in the listed paths; remove the entry"
    stale=$((stale + 1))
  fi
done

echo "${ignorefile}: ${count} entries, ${stale} no longer found"
exit 0
