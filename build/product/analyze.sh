#!/bin/bash
#set -eux
#===================================================================================================
#
# Analyze sources
#
#===================================================================================================
#---------------------------------------------------------------------------------------------------
# env
#---------------------------------------------------------------------------------------------------
dir_script="$(dirname $0)"
cd "$(cd ${dir_script}; cd ../..; pwd)" || exit 1

readonly DIR_BASE="$(pwd)"
. "${DIR_BASE}/build/env.properties"


#---------------------------------------------------------------------------------------------------
# check
#---------------------------------------------------------------------------------------------------
${DIR_BASE}/build/lib/check_installed.sh product-analyze
if [[ $? -ne 0 ]]; then
  exit 1
fi

#---------------------------------------------------------------------------------------------------
# prepare
#---------------------------------------------------------------------------------------------------
echo "$(basename $0)"
DIR_ANALYZE_DIST="${DIR_DIST}/.analyze"
if [[ -d "${DIR_ANALYZE_DIST}" ]]; then rm -fr "${DIR_ANALYZE_DIST}"; fi
mkdir -p "${DIR_ANALYZE_DIST}"

path_target="${DIR_ANALYZE_DIST}/target.lst"
path_report="${DIR_ANALYZE_DIST}/report.txt"

echo "  list sources"
pwd
find "${DIR_SRC}/bin" -type f                                                                      |
sed -e "s|${DIR_BASE}/||"                                                                          |
grep -v "\.DS_Store" >>"${path_target}"


#---------------------------------------------------------------------------------------------------
# analyze
#---------------------------------------------------------------------------------------------------
target_files=( $(cat "${path_target}") )
cmd=(
  shellcheck -x
    -e SC1090
    -e SC2086
    -e SC2155
    -e SC2164
    -e SC1117
    "${target_files[@]}"
)

echo -n '  '
echo "${cmd[@]}"
"${cmd[@]}" >"${path_report}"
retcode=$?

if [[ ${retcode} -ne 0 ]]; then
  count=$(cat "${path_report}" | grep -- "-- SC....: " | wc -l | sed -E 's|^ +||')
  cat "${path_report}"
  (
    echo "$(basename $0) failed."
    echo "  count: ${count}"
  ) >&2
  exit ${retcode}
fi


#---------------------------------------------------------------------------------------------------
# teardown
#---------------------------------------------------------------------------------------------------
if [[ -d "${DIR_ANALYZE_DIST}" ]]; then rm -fr "${DIR_ANALYZE_DIST}"; fi

echo "$(basename $0) success."
exit 0
