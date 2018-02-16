#!/bin/sh

dir_script="$(dirname $0)"
cd "$(cd ${dir_script}; cd ../..; pwd)" || exit 1

readonly DIR_BASE="$(pwd)"
. "${DIR_BASE}/build/env.properties"

product="stfw"
version="$(cat ${DIR_SRC}/VERSION)"
archive_name="${product}-${version}"
archive_name_with_dpends="${product}-with-depends-${version}"

if [ ! -f ${DIR_BASE}/dist/${archive_name_with_dpends}.tar.gz ]; then
  echo "build.sh を先に実行してください" 2>&1
  exit 2
fi

tar xzf ${DIR_BASE}/dist/${archive_name_with_dpends}.tar.gz -C ${DIR_BASE}/dist/
mv ${DIR_BASE}/dist/${archive_name_with_dpends} ${DIR_BASE}/dist/test
sudo docker-compose -f docker-compose-kcov.yml \
     run stfw-kcov --include-pattern /source \
         /source/dist/coverage \
         /source/test/ut/ut_all.sh /source/dist/test /source/dist/proj

rm -R ${DIR_BASE}/dist/test
rm -R ${DIR_BASE}/dist/proj

sudo docker-compose -f docker-compose-kcov.yml \
     run stfw-kcov --include-pattern /source \
         /source/dist/coverage \
         /source/build/product/integration_test.sh /source/dist/${archive_name_with_dpends}.tar.gz

