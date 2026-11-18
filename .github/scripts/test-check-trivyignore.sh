#!/usr/bin/env bash
# test-check-trivyignore.sh は除外リストの検査 (check-trivyignore.sh) と paths の照合 (trivyignore-lib.sh) を
# 見本の入力で確かめる (CI の ci.yml で PR ごとに実行。Docker 不要、yq v4 が必要)。
# レビューで見つかったすり抜け (ワイルドカードだけのパス・複数文書・引用符つきの日付など) を固定する。
#
#   使い方: test-check-trivyignore.sh   (リポジトリ直下で実行)
#
# 終了コード: 0 = すべて期待どおり / 1 = 期待と違うものがある
set -u

here="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=trivyignore-lib.sh
. "${here}/trivyignore-lib.sh"
check="${here}/check-trivyignore.sh"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# 今日を基準にした日付 (GNU date と BSD date の両方に対応)
day() { date -u -d "$1 days" +%F 2>/dev/null || date -u -v"$1"d +%F; }
tomorrow="$(day +1)"
ok="$(day +30)"
edge="$(day +60)"
far="$(day +61)"
today="$(date -u +%F)"
past="$(day -1)"

failures=0
expect() { # $1=名前 $2=期待する終了コード $3=ファイルの内容
  printf '%s' "$3" >"$tmp/case.yaml"
  bash "$check" "$tmp/case.yaml" >"$tmp/out.txt" 2>&1
  local rc=$?
  if [ "$rc" -eq "$2" ]; then
    echo "ok   $1"
  else
    echo "FAIL $1: exit ${rc}, want $2"
    sed 's/^/       /' "$tmp/out.txt"
    failures=$((failures + 1))
  fi
}

# entry は 1 行分の YAML を作る (引数で各項目を差し替える。空文字はその項目を書かない)
entry() { # $1=paths の値 $2=expired_at の値 $3=statement の値 $4=追加の行
  printf '  - id: CVE-2026-0001\n'
  [ -n "$1" ] && printf '    paths: %s\n' "$1"
  [ -n "$2" ] && printf '    expired_at: %s\n' "$2"
  [ -n "$3" ] && printf '    statement: %s\n' "$3"
  [ -n "${4:-}" ] && printf '%s\n' "$4"
  return 0
}
P='["opt/stfw/bundled/k6/*/k6"]'
S='waiting for a fix release'
doc() { printf 'vulnerabilities:\n%s' "$1"; }

echo "--- valid"
expect "repository file" 0 "$(cat .trivyignore.yaml)"
expect "30 days ahead" 0 "$(doc "$(entry "$P" "$ok" "$S")")"
expect "tomorrow" 0 "$(doc "$(entry "$P" "$tomorrow" "$S")")"
expect "exactly the limit (60 days)" 0 "$(doc "$(entry "$P" "$edge" "$S")")"
expect "literal path" 0 "$(doc "$(entry '["opt/stfw/bundled/k6/linux_amd64/k6"]' "$ok" "$S")")"
expect "empty list" 0 "vulnerabilities: []"

echo "--- paths"
expect "no paths" 1 "$(doc "$(entry "" "$ok" "$S")")"
expect "empty paths list" 1 "$(doc "$(entry "[]" "$ok" "$S")")"
expect "scalar paths" 1 "$(doc "$(entry '"opt/stfw/bundled/k6/*/k6"' "$ok" "$S")")"
expect "non-string path" 1 "$(doc "$(entry "[1]" "$ok" "$S")")"
for bad in '**' '?*/**' '{o,u}*/**' 'opt/**' 'opt/*' 'opt/stfw/bundled/k6/*' '*/stfw/bundled/k6' 'opt/*/bundled/k6' \
  'opt/stfw/linux_*/k6' 'opt/[s]tfw/x/k6' 'opt/stfw/k?/k6' '/opt/stfw/bundled/k6' 'opt/stfw//k6' 'opt/stfw/k6/'; do
  expect "path '${bad}'" 1 "$(doc "$(entry "[\"${bad}\"]" "$ok" "$S")")"
done
expect "one bad path among good ones" 1 "$(doc "$(entry '["opt/stfw/bundled/k6/*/k6", "opt/**"]' "$ok" "$S")")"

echo "--- expired_at"
expect "no expired_at" 1 "$(doc "$(entry "$P" "" "$S")")"
expect "quoted date" 1 "$(doc "$(entry "$P" "\"${ok}\"" "$S")")"
expect "today (Trivy already stops ignoring)" 1 "$(doc "$(entry "$P" "$today" "$S")")"
expect "yesterday" 1 "$(doc "$(entry "$P" "$past" "$S")")"
expect "61 days ahead" 1 "$(doc "$(entry "$P" "$far" "$S")")"
expect "slash format" 1 "$(doc "$(entry "$P" "${ok//-//}" "$S")")"
expect "date and time" 1 "$(doc "$(entry "$P" "${ok}T00:00:00Z" "$S")")"

echo "--- other fields"
expect "no statement" 1 "$(doc "$(entry "$P" "$ok" "")")"
expect "empty statement" 1 "$(doc "$(entry "$P" "$ok" '""')")"
expect "whitespace-only statement" 1 "$(doc "$(entry "$P" "$ok" '"   "')")"
expect "whitespace-only id" 1 "$(printf 'vulnerabilities:\n  - id: "  "\n    paths: %s\n    expired_at: %s\n    statement: x\n' "$P" "$ok")"
expect "no id" 1 "$(printf 'vulnerabilities:\n  - paths: %s\n    expired_at: %s\n    statement: x\n' "$P" "$ok")"
expect "unsupported entry key (purls)" 1 "$(doc "$(entry "$P" "$ok" "$S" '    purls: ["pkg:golang/x"]')")"

echo "--- file shape"
expect "other top-level key" 1 "$(doc "$(entry "$P" "$ok" "$S")")"$'\n'"misconfigurations:"$'\n'"  - id: AVD-DS-0001"
expect "second YAML document (bypass)" 1 "$(doc "$(entry "" "" "")")"$'\n'"---"$'\n'"{}"
expect "vulnerabilities is not a list" 1 "vulnerabilities: {}"
expect "entry is not a mapping" 1 "vulnerabilities: [x]"
expect "broken YAML" 1 "vulnerabilities: [unclosed"
expect "empty file" 1 ""

echo "--- path matching (same as Trivy: * matches one segment)"
match() { # $1=名前 $2=期待 (0=一致 / 1=不一致) $3=pattern $4=target
  ti_path_match "$3" "$4"
  local rc=$?
  if [ "$rc" -eq "$2" ]; then
    echo "ok   $1"
  else
    echo "FAIL $1: got ${rc}, want $2"
    failures=$((failures + 1))
  fi
}
K='opt/stfw/bundled/k6/*/k6'
match "amd64" 0 "$K" "opt/stfw/bundled/k6/linux_amd64/k6"
match "arm64" 0 "$K" "opt/stfw/bundled/k6/linux_arm64/k6"
match "missing segment" 1 "$K" "opt/stfw/bundled/k6/k6"
match "extra segment (* does not cross /)" 1 "$K" "opt/stfw/bundled/k6/a/b/k6"
match "other tool" 1 "$K" "opt/stfw/bundled/compare-files/linux_arm64/compare_files"
match "literal" 0 "opt/stfw/bundled/k6/linux_amd64/k6" "opt/stfw/bundled/k6/linux_amd64/k6"

echo
if [ "$failures" -ne 0 ]; then
  echo "${failures} case(s) failed"
  exit 1
fi
echo "all cases passed"
