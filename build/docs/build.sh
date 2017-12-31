#!/bin/bash
#set -eux
#===================================================================================================
#
# Build Documents
#
#===================================================================================================
#---------------------------------------------------------------------------------------------------
# env
#---------------------------------------------------------------------------------------------------
dir_script="$(dirname $0)"
cd "$(cd ${dir_script}; cd ../..; pwd)" || exit 1

DIR_BASE="$(pwd)"
DIR_SRC="${DIR_BASE}/docs/adoc"
DIR_DIST="${DIR_BASE}/docs"


#---------------------------------------------------------------------------------------------------
# prepare
#---------------------------------------------------------------------------------------------------
echo "init dist file"
if [[ -f "${DIR_DIST}/index.html" ]]; then rm -f "${DIR_DIST}/index.html"; fi


#---------------------------------------------------------------------------------------------------
# analyze
#---------------------------------------------------------------------------------------------------
build/docs/analyze.sh
retcode=$?
if [[ ${retcode} -ne 0 ]]; then exit ${retcode}; fi


#---------------------------------------------------------------------------------------------------
# generate
#---------------------------------------------------------------------------------------------------
echo "generate"

version="$(cat ${DIR_BASE}/src/VERSION)"
cmd=(
  asciidoctor
    --safe-mode "unsafe"
    -a "lang=ja"
    -b "html5"
    -d "book"
    --destination-dir "${DIR_DIST}"
    --attribute "source-highlighter=highlightjs"
    --attribute "linkcss"
    --attribute "stylesheet=readthedocs.css"
    --attribute "stylesdir=./stylesheets"
    --attribute "Version=${version}"
    --attribute "imagesdir=./images"
    "${DIR_SRC}/index.adoc"
)

echo -n '  '
echo "${cmd[@]}"
"${cmd[@]}"
retcode=$?

if [[ ${retcode} -ne 0 ]] || [[ ! -f "${DIR_DIST}/index.html" ]]; then
  echo "    error occured in acsiidoctor." >&2
  exit 1
fi


#---------------------------------------------------------------------------------------------------
# teardown
#---------------------------------------------------------------------------------------------------
echo ""
echo "build completed."
exit 0
