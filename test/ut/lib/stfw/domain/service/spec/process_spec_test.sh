#!/bin/bash

#load env
. ${STFW_HOME}/bin/lib/setenv
#target load
. ${DIR_BIN_LIB}/stfw/domain/service/spec/process_spec

# テスト対象の関数名 . はだめ _ は ok
# bashの?で落ちる場合、処理続行されない（検証できない）
function test_stfw_domain_service_spec_process_dirname() {
  local _expect="_seq_group_type"
  local _actual=$(stfw.domain.service.spec.process.dirname "type" "seq" "group")
  assertEquals ${_expect} ${_actual}
}

function test_stfw_domain_service_spec_process_dir() {
  local _expect="/exec/_seq_group_type"
  local _actual=$(stfw.domain.service.spec.process.dir "/exec" "type" "seq" "group")
  assertEquals ${_expect} ${_actual}
}

function test_stfw_domain_service_spec_get_process_type() {
  local _expect="type"
  local _actual=$(stfw.domain.service.spec.get_process_type "_seq_group_type")
  assertEquals ${_expect} ${_actual}
}

. ${SHUNIT}
