#!/bin/bash
#set -eux
#===================================================================================================
#
# Build Product
#
#===================================================================================================
#---------------------------------------------------------------------------------------------------
# env
#---------------------------------------------------------------------------------------------------
dir_script="$(dirname $0)"
cd "$(cd ${dir_script}; cd ../..; pwd)" || exit 1

DIR_BASE="$(pwd)"
DIR_SRC="${DIR_BASE}/src"
DIR_DIST="${DIR_BASE}/dist"


#---------------------------------------------------------------------------------------------------
# prepare
#---------------------------------------------------------------------------------------------------
echo "init dist directory"
if [[ -d "${DIR_DIST}" ]]; then rm -fr "${DIR_DIST}"; fi
mkdir -p "${DIR_DIST}"


#---------------------------------------------------------------------------------------------------
# analyze
#---------------------------------------------------------------------------------------------------
build/product/analyze.sh
retcode=$?
if [[ ${retcode} -ne 0 ]]; then exit ${retcode}; fi


#---------------------------------------------------------------------------------------------------
# package
#---------------------------------------------------------------------------------------------------
echo "package"

product="$(basename ${DIR_BASE})"
version="$(cat ${DIR_SRC}/VERSION)"

archive_name="${product}-${version}"
archive_name_with_dpends="${product}-with-depends-${version}"

echo "  copy sources"
dir_dist_work="${DIR_DIST}/${archive_name_with_dpends}"
mkdir "${dir_dist_work}"
cp -pr "${DIR_SRC}/"* "${dir_dist_work}/"

echo "  remove UT work files"
rm -fr "${dir_dist_work}"/archives
rm -fr "${dir_dist_work}"/modules

echo "  remove exclude files"
# shellcheck disable=SC2038
find "${dir_dist_work}" -type f -name ".gitkeep"  | xargs -I{} bash -c 'echo "rm -f {}"; rm -f {}'
# shellcheck disable=SC2038
find "${dir_dist_work}" -type f -name ".DS_Store" | xargs -I{} bash -c 'echo "rm -f {}"; rm -f {}'

echo "  run install script"
${dir_dist_work}/bin/install
retcode=$?
if [[ ${retcode} -ne 0 ]]; then echo "    error occured in install script." >&2; exit 1; fi

echo "  package with-depends-archive"
rm -fr "${dir_dist_work:?}/modules/"
cd ${DIR_DIST}
tar czf "./${archive_name_with_dpends}.tar.gz" "./${archive_name_with_dpends}"
retcode=$?
if [[ ${retcode} -ne 0 ]]; then echo "    error occured in tar." >&2; exit 1; fi
cd - >/dev/null 2>&1

echo "  package exclude-depends-archive"
rm -fr "${dir_dist_work:?}/archives"
cd ${DIR_DIST}
mv "./${archive_name_with_dpends}" "./${archive_name}"
tar czf "./${archive_name}.tar.gz" "./${archive_name}"
retcode=$?
if [[ ${retcode} -ne 0 ]]; then echo "    error occured in tar." >&2; exit 1; fi
cd - >/dev/null 2>&1

echo "  remove work files"
rm -fr "${DIR_DIST}/${archive_name:?}/"


#---------------------------------------------------------------------------------------------------
# test
#---------------------------------------------------------------------------------------------------
build/product/integration_test.sh "${DIR_DIST}/${archive_name_with_dpends}.tar.gz"
retcode=$?
if [[ ${retcode} -ne 0 ]]; then exit ${retcode}; fi


#---------------------------------------------------------------------------------------------------
# teardown
#---------------------------------------------------------------------------------------------------
echo "results:"
find "${DIR_DIST}" -type f

echo ""
echo "build completed."
exit 0
