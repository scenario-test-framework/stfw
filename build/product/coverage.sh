#!/bin/sh

dir_script="$(dirname $0)"
cd "$(cd ${dir_script}; cd ../..; pwd)" || exit 1

readonly DIR_BASE="$(pwd)"
. "${DIR_BASE}/build/env.properties"

product="stfw"
version="$(cat ${DIR_SRC}/VERSION)"
#archive_name="${product}-${version}"
archive_name_with_dpends="${product}-with-depends-${version}"

if [ ! -f ${DIR_BASE}/dist/${archive_name_with_dpends}.tar.gz ]; then
  echo "build.sh を先に実行してください" 2>&1
  exit 2
fi

# all UT
mkdir -p ${DIR_BASE}/dist/test/proj
tar xzf ${DIR_BASE}/dist/${archive_name_with_dpends}.tar.gz -C ${DIR_BASE}/dist/
mv ${DIR_BASE}/dist/${archive_name_with_dpends} ${DIR_BASE}/dist/test/stfw
docker-compose                                                                                     \
  --file docker-compose-kcov.yml                                                                   \
  run stfw-kcov                                                                                    \
    --include-pattern /source /source/dist/coverage                                                \
    /source/test/ut/ut_all.sh /source/dist/test/stfw /source/dist/test/proj
# container削除
docker-compose -f docker-compose-kcov.yml down

# IT
rm -R ${DIR_BASE}/dist/test/
docker-compose                                                                                     \
  --file docker-compose-kcov.yml                                                                   \
  run stfw-kcov                                                                                    \
    --include-pattern /source /source/dist/coverage                                                \
    /source/build/product/integration_test.sh /source/dist/${archive_name_with_dpends}.tar.gz
# container削除
docker-compose -f docker-compose-kcov.yml down

# 結果表示
covered=$(
  cat dist/coverage/index.json                                                                     | # 結果データから
  sed -e 's|^var data = ||'                                                                        | # dataオブジェクトに絞り込み
  sed -e 's|^\]\};|]}|'                                                                            |
  grep -v "^var "                                                                                  |
  tr '\n' ' '                                                                                      | # 1line化
  sed -e 's|.*merged_files:\[||'                                                                   | # merged_filesオブジェクトに絞り込み
  sed -e 's|, \]\} $||'                                                                            |
  jq '.covered'                                                                                    | # カバレッジを抽出
  sed -e 's|^"||' -e 's|"$||'                                                                        # アンクォート
)
echo "covered    : ${covered}%"
echo "html report: ${DIR_BASE}/dist/coverage/index.html"
