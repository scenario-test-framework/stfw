#!/bin/bash
. ${DIR_UT}/test_utils

function oneTimeSetUp() {
  init_stfw_proj "process_initialized"
  export STFW_PROJ_DIR="$(get_stfw_proj_dir)"

  # load env
  . ${STFW_HOME}/bin/lib/setenv
  # target load
  . ${DIR_BIN_LIB}/stfw/domain/service/spec/process_spec
}

function oneTimeTearDown() {
  clean_stfw_proj
}


function test_plugin_installed_dir() {
  local _actual=$(stfw.domain.service.spec.process.plugin_installed_dir "NotExist")
  assertNull "${_actual}"

  local _actual=$(stfw.domain.service.spec.process.plugin_installed_dir "scripts")
  assertEquals "${STFW_HOME}/plugins/process/scripts" "${_actual}"
}


function test_is_installed() {
  local _plugin_path=$(stfw.domain.service.spec.process.plugin_installed_dir "scripts")
  local _actual=$(stfw.domain.service.spec.process.is_installed "${_plugin_path}")
  assertEquals "true" "${_actual}"
}


function test_dirname() {
  local _actual=$(stfw.domain.service.spec.process.dirname "TypeName" "99" "GroupA")
  assertEquals "_99_GroupA_TypeName" "${_actual}"
}


function test_dir() {
  local _actual=$(stfw.domain.service.spec.process.dir "/path/to" "TypeName" "99" "GroupA")
  assertEquals "/path/to/_99_GroupA_TypeName" "${_actual}"
}


function test_type() {
  local _actual=$(stfw.domain.service.spec.process.type "_99_GroupA_TypeName")
  assertEquals "TypeName" "${_actual}"
}


. ${SHUNIT}
