#!/usr/bin/env bash
# check-trivyignore.sh は Trivy の除外リスト (.trivyignore.yaml) が決まりを守っているかを検査する
# (AS-BUILT §10.4)。全体の除外・無期限の除外・理由の無い除外を防ぐ。
#
#   使い方: check-trivyignore.sh [<ignore-file>]   (既定: .trivyignore.yaml。yq v4 が必要)
#
# 決まり:
#   - ファイルは YAML の文書 1 つで、最上位は vulnerabilities (一覧) だけ
#     (Trivy は最初の文書しか読まないため、2 つ目以降の文書で検査をすり抜けられないようにする)
#   - 各行のキーは id / paths / expired_at / statement だけで、すべて必須
#   - paths は一覧で、各要素は trivyignore-lib.sh の書き方 (先頭 2 階層とファイル名を必ず書く)
#   - expired_at は引用符なしの yyyy-mm-dd で、明日以降かつ今日から MAX_DAYS 日以内
#     (Trivy は期限日の 0 時 (UTC) に除外をやめるため、期限日当日はもう効かない)
#
# 終了コード: 0 = 決まりどおり / 1 = 違反あり (yq の失敗を含む)
set -u

here="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=trivyignore-lib.sh
. "${here}/trivyignore-lib.sh"

file="${1:-.trivyignore.yaml}"
max_days="${TRIVYIGNORE_MAX_DAYS:-60}"

die() {
  echo "::error file=${file}::$1"
  exit 1
}

command -v yq >/dev/null 2>&1 || die "yq (v4) is required to check the ignore list"
[ -f "$file" ] || die "file not found"

# q は yq の結果を返す。yq が失敗したら記録し (q は $(...) の中で呼ぶため exit では止まらない)、
# 呼び出し側の判定が空の値で通ってしまっても、最後に必ず検査全体を失敗にする
errlog="$(mktemp)"
trap 'rm -f "$errlog"' EXIT
q() {
  local out
  if ! out="$(yq "$1" "$file" 2>&1)"; then
    echo "::error file=${file}::cannot read the file with yq (${1}): ${out}" >>"$errlog"
    return 1
  fi
  printf '%s' "$out"
}
yq_failed() {
  if [ -s "$errlog" ]; then
    cat "$errlog"
    return 0
  fi
  return 1
}

docs="$(q 'document_index' | sort -u | grep -c .)"
yq_failed && exit 1
[ "$docs" = "1" ] || die "must contain exactly one YAML document (found ${docs})"
[ "$(q 'tag')" = "!!map" ] || die "top level must be a mapping with vulnerabilities"
others="$(q 'keys | map(select(. != "vulnerabilities")) | join(",")')"
[ -z "$others" ] || die "unsupported top-level keys: ${others} (only vulnerabilities is allowed)"
[ "$(q '.vulnerabilities | tag')" = "!!seq" ] || die "vulnerabilities must be a list"
count="$(q '.vulnerabilities | length')"
yq_failed && exit 1
[[ "$count" =~ ^[0-9]+$ ]] || die "cannot count entries (got '${count}')"

today="$(date -u +%F)"
# GNU date (GitHub Actions) と BSD date (macOS) の両方で日付を足す
limit="$(date -u -d "+${max_days} days" +%F 2>/dev/null || date -u -v+"${max_days}"d +%F)"

rc=0
fail() {
  echo "::error file=${file}::$1"
  rc=1
}

for ((i = 0; i < count; i++)); do
  e=".vulnerabilities[$i]"
  if [ "$(q "$e | tag")" != "!!map" ]; then
    fail "entry $((i + 1)): must be a mapping"
    continue
  fi
  id="$(q "$e.id // \"\"")"
  label="entry $((i + 1)) (${id:-no id})"
  extra="$(q "$e | keys | map(select(. != \"id\" and . != \"paths\" and . != \"expired_at\" and . != \"statement\")) | join(\",\")")"
  [ -z "$extra" ] || fail "${label}: unsupported keys: ${extra} (use id, paths, expired_at and statement)"
  if [ "$(q "$e.id | tag")" != "!!str" ] || [ -z "${id//[[:space:]]/}" ]; then
    fail "${label}: id is required"
  fi

  if [ "$(q "$e.paths | tag")" != "!!seq" ] || [ "$(q "$e.paths | length")" = "0" ]; then
    fail "${label}: paths is required as a list (without it the entry applies to the whole image)"
  else
    npaths="$(q "$e.paths | length")"
    for ((j = 0; j < npaths; j++)); do
      if [ "$(q "$e.paths[$j] | tag")" != "!!str" ]; then
        fail "${label}: paths[$j] must be a string"
        continue
      fi
      p="$(q "$e.paths[$j]")"
      problem="$(ti_path_problem "$p")"
      [ -z "$problem" ] || fail "${label}: path '${p}': ${problem}"
    done
  fi

  exp="$(q "$e.expired_at // \"\"")"
  if [ "$(q "$e.expired_at | tag")" != "!!timestamp" ] || ! [[ "$exp" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; then
    fail "${label}: expired_at is required as an unquoted yyyy-mm-dd date (got '${exp}')"
  elif ! [[ "$exp" > "$today" ]]; then
    fail "${label}: expired on ${exp} (Trivy stops ignoring on that day); adopt the fix, or extend it with a reason"
  elif [[ "$exp" > "$limit" ]]; then
    fail "${label}: expired_at ${exp} is more than ${max_days} days ahead (limit ${limit})"
  fi

  stmt="$(q "$e.statement // \"\"")"
  # 空白だけの理由は書いていないのと同じ
  if [ "$(q "$e.statement | tag")" != "!!str" ] || [ -z "${stmt//[[:space:]]/}" ]; then
    fail "${label}: statement is required (what are we waiting for?)"
  fi
done

yq_failed && rc=1
if [ "$rc" -eq 0 ]; then
  echo "${file}: ${count} entries, all valid (expiry from tomorrow up to ${limit})"
fi
exit "$rc"
