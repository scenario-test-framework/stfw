#!/bin/bash
#set -eux
#===================================================================================================
#
# Publish Documents
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

version=$(cat "${PATH_VERSION}")


#---------------------------------------------------------------------------------------------------
# check
#---------------------------------------------------------------------------------------------------
if [[ "${GITHUB_TOKEN}x" = "x" ]]; then
  echo "GITHUB_TOKEN is not defined." >&2
  exit 1
fi


#-------------------------------------------------------------------------------
# オプション解析
#-------------------------------------------------------------------------------
while :; do
  case $1 in
    --)
      shift
      break
      ;;

    *)
      break
      ;;
  esac
done


#---------------------------------------------------------------------------------------------------
# main
#---------------------------------------------------------------------------------------------------
echo "$(basename $0)"
retcode=0

add_git_config

echo "  update"
git pull
git checkout "${BRANCH_GHPAGES}"
exit_on_fail "git checkout \"${BRANCH_GHPAGES}\"" $?
git reset
exit_on_fail "git reset" $?

echo "  clear"
rm -f ./index.html
rm -fr ./stylesheets/
rm -fr ./images/

echo "  move"
if [[ -f ./docs/index.html   ]]; then mv ./docs/index.html ./;   fi
if [[ -d ./docs/stylesheets/ ]]; then mv ./docs/stylesheets/ ./; fi
if [[ -d ./docs/images/      ]]; then mv ./docs/images/ ./;      fi

echo "  staging"
if [[ -f ./index.html   ]]; then git add ./index.html;   fi
if [[ -d ./stylesheets/ ]]; then git add ./stylesheets/; fi
if [[ -d ./images/      ]]; then git add ./images/;      fi

echo "  commit"
git commit -m "chore(release docs): v${version}"
exit_on_fail "commit" $?

echo "  push"
git push origin "${BRANCH_GHPAGES}"
exit_on_fail "push" $?

echo "  reset"
git reset --hard
git clean -df
git checkout "${BRANCH_MASTER}"
exit_on_fail "reset" $?


#---------------------------------------------------------------------------------------------------
# teardown
#---------------------------------------------------------------------------------------------------
if [[ ${retcode} -eq 0 ]]; then
  echo "$(basename $0) success."
  exitcode=0
else
  echo "$(basename $0) failed." >&2
  exitcode=1
fi

remove_credential
exit ${exitcode}
