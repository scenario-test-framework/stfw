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

sudo docker-compose -f docker-compose-kcov.yml up

