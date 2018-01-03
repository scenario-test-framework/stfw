#!/bin/bash
#set -eux
#===================================================================================================
#
# Pre Release
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
# functions
#---------------------------------------------------------------------------------------------------
#-------------------------------------------------------------------------------
# version更新
#-------------------------------------------------------------------------------
function update_version() {
  echo "${FUNCNAME[0]}"

  local _cur_version=$(cat "${PATH_VERSION}")
  local _release_version="${_cur_version//-SNAPSHOT/}"
  local _release_tag="v${_release_version}"
  local _commit_message="chore(release): ${_release_tag}"

  echo "  update version file"
  echo "    ${_cur_version} -> ${_release_version}"
  echo "${_release_version}" >"${PATH_VERSION}"

# TODO commitizen
#  echo "  generate changelog"
#  exit_on_fail "generate changelog" $?

  add_git_config

  echo "  git add"
  git add --all .

  echo "  git commit"
  git commit -m "${_commit_message}"
  exit_on_fail "git commit" $?

  echo "  git tag"
  git tag -a "${_release_tag}" -m "${_commit_message}"
  exit_on_fail "git tag" $?

  echo "  git push branch ${BRANCH_MASTER}"
  git push origin "${BRANCH_MASTER}"
  exit_on_fail "git push branch ${BRANCH_MASTER}" $?

  echo "  git push tag ${_release_tag}"
  git push origin "${_release_tag}"
  exit_on_fail "git push tag ${_release_tag}" $?

  return 0
}


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

echo ""
${DIR_BUILD}/product/build.sh
exit_on_fail "build product" $?

echo ""
${DIR_BUILD}/docs/build.sh
exit_on_fail "build docs" $?

echo ""
update_version


#---------------------------------------------------------------------------------------------------
# teardown
#---------------------------------------------------------------------------------------------------
echo "$(basename $0) success."
exit 0
