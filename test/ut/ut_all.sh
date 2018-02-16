#!/bin/bash
# example:  ./ut_all.sh ~/project/stfw/src/ /tmp/proj

BASEDIR=$(dirname $0)
export STFW_HOME=${1:?}
export STFW_PROJ_DIR=${2:?}
export SHUNIT=${BASEDIR}/shunit2 

find ${BASEDIR}/lib -type f |
while read FILE
do
  echo "--------------------------------------------------------------------------"
  echo " shunit: ${FILE}"
  echo "--------------------------------------------------------------------------"
  ${FILE}
done
