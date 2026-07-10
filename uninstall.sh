#!/usr/bin/env bash
set -euo pipefail

binary_name="stfw"

usage() {
  cat <<'EOF'
Usage:
  ./uninstall.sh [bindir]
  ./uninstall.sh --bindir /usr/local/bin

Environment variables:
  STFW_BINDIR   Directory containing the stfw binary.
EOF
}

can_write_dir() {
  local dir parent
  dir="$1"
  while [ ! -d "${dir}" ]; do
    parent="$(dirname "${dir}")"
    if [ "${parent}" = "${dir}" ]; then
      return 1
    fi
    dir="${parent}"
  done
  [ -w "${dir}" ]
}

default_bindir() {
  if [ -x "/usr/local/bin/${binary_name}" ]; then
    printf '%s\n' "/usr/local/bin"
    return
  fi
  printf '%s\n' "${HOME}/.local/bin"
}

remove_binary() {
  local target="$1"

  if [ ! -e "${target}" ]; then
    echo "uninstall.sh: ${target} does not exist" >&2
    exit 1
  fi

  if can_write_dir "$(dirname "${target}")"; then
    rm -f "${target}"
    return
  fi

  if command -v sudo >/dev/null 2>&1; then
    sudo rm -f "${target}"
    return
  fi

  echo "uninstall.sh: cannot remove ${target} and sudo is not available" >&2
  exit 1
}

bindir="${STFW_BINDIR:-$(default_bindir)}"
bindir_set=0

while [ "$#" -gt 0 ]; do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    --bindir)
      shift
      if [ "$#" -eq 0 ]; then
        echo "uninstall.sh: --bindir requires a value" >&2
        exit 1
      fi
      bindir="$1"
      bindir_set=1
      ;;
    *)
      if [ "${bindir_set}" -eq 0 ]; then
        bindir="$1"
        bindir_set=1
      else
        echo "uninstall.sh: unexpected argument: $1" >&2
        usage >&2
        exit 1
      fi
      ;;
  esac
  shift
done

target="${bindir}/${binary_name}"
remove_binary "${target}"
echo "uninstall.sh: removed ${target}"
echo "uninstall.sh: if needed, remove ${bindir} from your PATH manually"
