#!/bin/bash
#set -eux
#===================================================================================================
#
# Post Release
#
# env
#   GITHUB_TOKEN
#
#===================================================================================================
#---------------------------------------------------------------------------------------------------
# env
#---------------------------------------------------------------------------------------------------
dir_script="$(dirname $0)"
cd "$(cd ${dir_script}; cd ../..; pwd)" || exit 1

readonly DIR_BASE="$(pwd)"
. "${DIR_BASE}/build/env.properties"
. "${DIR_BUILD_LIB}/common.sh"


#---------------------------------------------------------------------------------------------------
# check
#---------------------------------------------------------------------------------------------------
if [[ "${GITHUB_TOKEN}x" = "x" ]]; then
  echo "GITHUB_TOKEN is not defined." >&2
  exit 1
fi


#---------------------------------------------------------------------------------------------------
# main
#---------------------------------------------------------------------------------------------------
echo "$(basename $0)"

echo "  update version file"
released_version=$(cat "${PATH_VERSION}")
# shellcheck disable=SC2034
next_version=$(
  echo ${released_version}                                                                         |
  ( IFS=".$IFS" ; read major minor bugfix && echo ${major}.$(( minor + 1 )).0-SNAPSHOT )
)

echo "    ${released_version} -> ${next_version}"
echo "${next_version}" >"${PATH_VERSION}"

add_git_config

echo "  git add"
git add --all .

echo "  git commit"
git commit -m "chore(VERSION): start ${next_version}"
exit_on_fail "git commit" $?

echo "  git push branch ${BRANCH_MASTER}"
git push origin "${BRANCH_MASTER}"
exit_on_fail "git push branch ${BRANCH_MASTER}" $?


#---------------------------------------------------------------------------------------------------
# teardown
#---------------------------------------------------------------------------------------------------
echo "$(basename $0) success."
exit 0
