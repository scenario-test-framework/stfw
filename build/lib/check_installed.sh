#!/bin/bash
dir_script="$(dirname $0)"
cd "$(cd ${dir_script}; cd ../..; pwd)" || exit 1

readonly DIR_BASE="$(pwd)"
. "${DIR_BASE}/build/env.properties"

# 前提チェック
which serverspec-init > /dev/null
if [[ $? -ne 0 ]]; then
  echo "serverspec is not installed." >&2
  exit 1
fi

# 指定されたターゲットのチェック
( cd ${DIR_BUILD}/lib/serverspec; rake spec:$1 ) >&2
if [[ $? -ne 0 ]]; then
  exit 1
fi
