#!/bin/bash
#set -eux
#===================================================================================================
#
# packaged archive integration test
#
#===================================================================================================
#---------------------------------------------------------------------------------------------------
# env
#---------------------------------------------------------------------------------------------------
dir_script="$(dirname $0)"
cd "$(cd ${dir_script}; cd ..; pwd)" || exit 6

DIR_BASE="$(pwd)"
DIR_DIST="${DIR_BASE}/dist"

DIR_TEST_WORK="${DIR_DIST}/test"
if [[ ! -d "${DIR_TEST_WORK}" ]]; then mkdir -p "${DIR_TEST_WORK}"; fi


path_archive="$1"


#---------------------------------------------------------------------------------------------------
# prepare
#---------------------------------------------------------------------------------------------------
echo "integration test"
echo "  prepare"
DIR_TEST_HOME="${DIR_TEST_WORK}/stfw"
DIR_TEST_PROJ="${DIR_TEST_WORK}/proj"

# 配布アーカイブ展開
echo "    extract package"
if [[ -d "${DIR_TEST_HOME}" ]]; then rm -fr "${DIR_TEST_HOME}"; fi
mkdir -p "${DIR_TEST_HOME}"
cd "${DIR_TEST_HOME}"
tar xzf "${path_archive}"
mv ./stfw-*/* .
rm -fr ./stfw-*

# install
echo "    install"
"${DIR_TEST_HOME}/bin/install"

# PATH追加
echo "    add path"
export PATH="${DIR_TEST_HOME}/bin:${PATH}"


#-------------------------------------------------------------------------------
# project init
#-------------------------------------------------------------------------------
STEP="project init"
echo "    ${STEP}"

if [[ -d "${DIR_TEST_PROJ}" ]]; then rm -fr "${DIR_TEST_PROJ}"; fi
mkdir -p "${DIR_TEST_PROJ}"
cd "${DIR_TEST_PROJ}"
stfw init
retcode=$?
if [[ ${retcode} -ne 0 ]]; then echo "      error occurred in test.${STEP} step." >&2; return 1; fi


#-------------------------------------------------------------------------------
# create scenario
#-------------------------------------------------------------------------------
STEP="create scenario"
echo "    ${STEP}"

# scenario init
cd "${DIR_TEST_PROJ}/scenario"
stfw scenario -i test
retcode=$?
if [[ ${retcode} -ne 0 ]]; then echo "      error occurred in test.${STEP} step." >&2; return 1; fi

# bizdate init (day1)
cd "${DIR_TEST_PROJ}/scenario/test"
stfw bizdate -i 10 99990101
retcode=$?
if [[ ${retcode} -ne 0 ]]; then echo "      error occurred in test.${STEP} step." >&2; return 1; fi

# process-scripts init
cd "${DIR_TEST_PROJ}/scenario/test/_10_99990101"
stfw process -i 10 pre scripts
retcode=$?
if [[ ${retcode} -ne 0 ]]; then echo "      error occurred in test.${STEP} step." >&2; return 1; fi

# bizdate init (day2)
cd "${DIR_TEST_PROJ}/scenario/test"
stfw bizdate -i 20 99990102
retcode=$?
if [[ ${retcode} -ne 0 ]]; then echo "      error occurred in test.${STEP} step." >&2; return 1; fi

# process-scripts init
cd "${DIR_TEST_PROJ}/scenario/test/_20_99990102"
stfw process -i 10 pre scripts
retcode=$?
if [[ ${retcode} -ne 0 ]]; then echo "      error occurred in test.${STEP} step." >&2; return 1; fi

# scenario gen-dig
cd "${DIR_TEST_PROJ}/scenario/test"
stfw scenario -G
retcode=$?
if [[ ${retcode} -ne 0 ]]; then echo "      error occurred in test.${STEP} step." >&2; return 1; fi


#-------------------------------------------------------------------------------
# run scenario
#-------------------------------------------------------------------------------
STEP="run scenario"
echo "    ${STEP}"

# server start
stfw server start
retcode=$?
if [[ ${retcode} -ne 0 ]]; then echo "      error occurred in test.${STEP} step." >&2; return 1; fi

stfw server status
retcode=$?
if [[ ${retcode} -ne 0 ]]; then echo "      error occurred in test.${STEP} step." >&2; return 1; fi

# run scenario
stfw run -f test
retcode=$?
if [[ ${retcode} -ne 0 ]]; then echo "      error occurred in test.${STEP} step." >&2; return 1; fi

# server stop
stfw server stop
retcode=$?
if [[ ${retcode} -ne 0 ]]; then echo "      error occurred in test.${STEP} step." >&2; return 1; fi

stfw server status
retcode=$?
if [[ ${retcode} -ne 0 ]]; then echo "      error occurred in test.${STEP} step." >&2; return 1; fi


#---------------------------------------------------------------------------------------------------
# teardown
#---------------------------------------------------------------------------------------------------
if [[ -d "${DIR_TEST_WORK}" ]]; then rm -fr "${DIR_TEST_WORK}"; fi

echo "  integration test success."
exit 0
