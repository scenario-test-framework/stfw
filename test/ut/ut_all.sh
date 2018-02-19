#!/bin/bash
dir_script="$(dirname $0)"
cd "$(cd ${dir_script}; pwd)"


DIR_SCM_ROOT=$(cd ../..; pwd)
export STFW_HOME="${DIR_SCM_ROOT}/src"
export DIR_UT="${DIR_SCM_ROOT}/test/ut"
export SHUNIT="${DIR_UT}/shunit2"


PATH_RETCODES="/tmp/$(basename $0 .sh).retcodes"
trap '
  echo "detect signal."
  rm -f ${PATH_RETCODES}
  exit 1
' SIGINT SIGQUIT SIGTERM


find "${DIR_UT}/scripts" -type f                                                                   |
sort                                                                                               |
while IFS= read cur_path; do
  cur_relpath=$(echo "${cur_path}" | sed -e "s|${DIR_UT}/||g")
  echo "----------------------------------------------------------------------------------------------------"
  echo " shunit: ${cur_relpath}"
  echo "----------------------------------------------------------------------------------------------------"
  "${cur_path}"
  echo "$?" >>"${PATH_RETCODES}"
done


error_retcodes=$(sed -e 's|0||g' <"${PATH_RETCODES}")
rm -f "${PATH_RETCODES}"
if [[ "${error_retcodes}x" != "x" ]]; then
  exit 1
fi
exit 0
