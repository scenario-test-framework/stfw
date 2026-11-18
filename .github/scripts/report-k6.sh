#!/usr/bin/env bash
# report-k6.sh は、全部入り版に同梱した k6 の「修正版がある HIGH 未満の脆弱性」を報告する
# (合否には含めない。AS-BUILT §10.4)。k6 は他社製で、部品の修正版が出ても k6 の新しい版が
# 出るまで直せないため、件数を実行結果の画面に注記し、明細を実行の概要に書くだけにする。
#
#   使い方: TRIVY_IMAGE=<image@digest> report-k6.sh <full-image>
#
# 常に終了コード 0 を返す (Trivy の失敗も警告の注記にとどめ、定期チェックを失敗させない)。
set -u

: "${TRIVY_IMAGE:?TRIVY_IMAGE is required}"
if [ "$#" -ne 1 ]; then
  echo "usage: report-k6.sh <full-image>" >&2
  exit 0
fi
full="$1"
cache="${TRIVY_CACHE_DIR:-${RUNNER_TEMP:-/tmp}/trivy-cache}"
mkdir -p "$cache"

if ! json="$(docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v "$cache:/root/.cache/" "$TRIVY_IMAGE" image \
  --scanners vuln --pkg-types library --ignore-unfixed --severity UNKNOWN,LOW,MEDIUM \
  --skip-dirs opt/stfw/bundled/compare-files --skip-dirs opt/stfw/bundled/logfilter \
  --skip-files usr/local/bin/stfw \
  --format json --no-progress --quiet "$full")"; then
  echo "::warning::bundled k6: informational scan failed"
  exit 0
fi

vulns='[.Results[]? | .Vulnerabilities[]?]'
count="$(jq "$vulns | length" <<<"$json")"
by_sev="$(jq -r "$vulns | group_by(.Severity) | map(\"\(.[0].Severity) \(length)\") | join(\", \")" <<<"$json")"
rows="$(jq -r "$vulns | .[] | \"| \(.VulnerabilityID) | \(.Severity) | \(.PkgName) | \(.InstalledVersion) | \(.FixedVersion) |\"" <<<"$json")"

summary="${GITHUB_STEP_SUMMARY:-/dev/stdout}"
{
  echo "### Bundled k6: fixable findings below HIGH (not part of the gate)"
  echo
  if [ "$count" -eq 0 ]; then
    echo "None."
  else
    echo "${count} findings (${by_sev}). They are fixed when a k6 release includes the dependency fixes."
    echo
    echo "| ID | Severity | Package | Installed | Fixed |"
    echo "|---|---|---|---|---|"
    echo "$rows"
  fi
} >>"$summary"

if [ "$count" -gt 0 ]; then
  echo "::warning::bundled k6 has ${count} fixable findings below HIGH (${by_sev}); see the run summary"
else
  echo "bundled k6: no fixable findings below HIGH"
fi
exit 0
