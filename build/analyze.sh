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
cd "$(cd ${dir_script}; cd ..; pwd)" || exit 6

DIR_BASE="$(pwd)"
DIR_SRC="${DIR_BASE}/src"
DIR_DIST="${DIR_BASE}/dist"

DIR_ANALYZE_DIST="${DIR_DIST}/analyze"
if [[ ! -d "${DIR_ANALYZE_DIST}" ]]; then mkdir -p "${DIR_ANALYZE_DIST}"; fi

path_target="${DIR_ANALYZE_DIST}/target.lst"
path_report="${DIR_ANALYZE_DIST}/report.txt"


#---------------------------------------------------------------------------------------------------
# check
#---------------------------------------------------------------------------------------------------
if [[ "$(which shellcheck)x" = "x" ]]; then
  echo "shellcheck is not installed." >&2
  exit 1
fi


#---------------------------------------------------------------------------------------------------
# prepare
#---------------------------------------------------------------------------------------------------
echo "analyze"
echo "  list sources"
find "${DIR_SRC}/bin" -type f                                                                      |
grep -v "lib/binary/"                                                                              |
grep -v "lib/Tukubai/"                                                                             |
grep -v "lib/Parsrs/"                                                                              |
grep -v "lib/yaml2json"                                                                            |
grep -v "lib/json2yaml"                                                                            |
grep -v "\.DS_Store" >>"${path_target}"


#---------------------------------------------------------------------------------------------------
# analyze
#---------------------------------------------------------------------------------------------------
echo "  run shellcheck"
target_files=( $(cat "${path_target}") )
shellcheck -x                                                                                      \
  -e SC1090                                                                                        \
  -e SC2086                                                                                        \
  -e SC2155                                                                                        \
  -e SC2164                                                                                        \
  "${target_files[@]}" >"${path_report}"
retcode=$?
if [[ ${retcode} -ne 0 ]]; then
  count=$(cat "${path_report}" | grep -- "-- SC....: " | wc -l)

  cat "${path_report}"
  (
    echo "  analyze failed."
    echo "    count: ${count}"
  ) >&2
  exit ${retcode}
fi


#---------------------------------------------------------------------------------------------------
# teardown
#---------------------------------------------------------------------------------------------------
if [[ -d "${DIR_ANALYZE_DIST}" ]]; then rm -fr "${DIR_ANALYZE_DIST}"; fi

echo "  analyze success."
exit 0
