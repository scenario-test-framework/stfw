#!/usr/bin/env bash
# trivyignore-lib.sh は除外リスト (.trivyignore.yaml) の paths の決まりと照合を、検査スクリプトと
# 消し忘れ検査で共有する (source して使う。AS-BUILT §10.4)。
#
# paths の書き方 (Trivy の照合 = doublestar のうち、範囲を限定できる書き方だけを許す):
#   - / 区切りの 3 階層以上。先頭 2 階層と最後 (ファイル名) はワイルドカードを含まない
#   - 途中の階層だけ、階層全体を * にできる (例: opt/stfw/bundled/k6/*/k6)。* は 1 階層だけに一致する
#   - **、?、[ ]、{ }、\ と、文字と * を混ぜた階層 (例: linux_*) は使わない (広く一致しうるため)

# ti_path_problem は paths の 1 つが決まりを守っていなければ理由を出力する (守っていれば何も出力しない)。
ti_path_problem() {
  local p="$1" seg i n
  if [ -z "$p" ]; then
    echo "empty path"
    return
  fi
  case "$p" in
    *'**'* | *'?'* | *'['* | *']'* | *'{'* | *'}'* | *\\*)
      echo "uses **, ?, [ ], { } or \\ (only a whole-segment * is allowed)"
      return
      ;;
    /* | */ | *//*)
      echo "must be a relative path without empty segments"
      return
      ;;
  esac
  local -a segs
  IFS=/ read -r -a segs <<<"$p"
  n="${#segs[@]}"
  if [ "$n" -lt 3 ]; then
    echo "needs at least 3 segments (name the directory and the file)"
    return
  fi
  for ((i = 0; i < n; i++)); do
    seg="${segs[$i]}"
    if [ "$seg" = "*" ]; then
      if [ "$i" -lt 2 ] || [ "$i" -eq $((n - 1)) ]; then
        echo "the first two segments and the file name must not be *"
        return
      fi
    elif ! [[ "$seg" =~ ^[A-Za-z0-9._+-]+$ ]]; then
      echo "segment '${seg}' must be a literal name or exactly *"
      return
    fi
  done
}

# ti_path_match は、決まりどおりの pattern が target (Trivy の Target) に一致すれば 0 を返す。
# * は 1 階層だけに一致する (Trivy の doublestar と同じ)。
ti_path_match() {
  local -a ps ts
  local i
  IFS=/ read -r -a ps <<<"$1"
  IFS=/ read -r -a ts <<<"$2"
  [ "${#ps[@]}" -eq "${#ts[@]}" ] || return 1
  for ((i = 0; i < ${#ps[@]}; i++)); do
    if [ "${ps[$i]}" = "*" ]; then
      [ -n "${ts[$i]}" ] || return 1
    elif [ "${ps[$i]}" != "${ts[$i]}" ]; then
      return 1
    fi
  done
  return 0
}
