#!/usr/bin/env bash
set -euo pipefail

repo="scenario-test-framework/stfw"
binary_name="stfw"

usage() {
  cat <<'EOF'
Usage:
  ./install.sh [version] [bindir]
  ./install.sh --version X.Y.Z --bindir /usr/local/bin

Environment variables:
  STFW_VERSION  Version to install. Accepts X.Y.Z, vX.Y.Z, or latest.
  STFW_BINDIR   Directory to install the stfw binary into.
EOF
}

need_cmd() {
  local cmd
  for cmd in "$@"; do
    if ! command -v "${cmd}" >/dev/null 2>&1; then
      echo "install.sh: ${cmd} is required" >&2
      exit 1
    fi
  done
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
  if can_write_dir "/usr/local/bin" || command -v sudo >/dev/null 2>&1; then
    printf '%s\n' "/usr/local/bin"
    return
  fi
  printf '%s\n' "${HOME}/.local/bin"
}

resolve_latest_tag() {
  local latest_url latest_tag
  latest_url="$(curl -fsSIL -o /dev/null -w '%{url_effective}' "https://github.com/${repo}/releases/latest")"
  latest_tag="${latest_url##*/}"
  if [ -z "${latest_tag}" ] || [ "${latest_tag}" = "latest" ]; then
    echo "install.sh: failed to resolve latest release tag" >&2
    exit 1
  fi
  printf '%s\n' "${latest_tag}"
}

normalize_version() {
  local input="$1"
  if [ "${input}" = "latest" ]; then
    resolved_tag="$(resolve_latest_tag)"
    resolved_version="${resolved_tag#v}"
    return
  fi

  if [ "${input#v}" != "${input}" ]; then
    resolved_tag="${input}"
    resolved_version="${input#v}"
    return
  fi

  resolved_tag="v${input}"
  resolved_version="${input}"
}

install_binary() {
  local src="$1" dest_dir="$2" dest_path="$2/$binary_name"

  if can_write_dir "${dest_dir}"; then
    mkdir -p "${dest_dir}"
    install -m 0755 "${src}" "${dest_path}"
    return
  fi

  if command -v sudo >/dev/null 2>&1; then
    sudo mkdir -p "${dest_dir}"
    sudo install -m 0755 "${src}" "${dest_path}"
    return
  fi

  echo "install.sh: cannot write to ${dest_dir} and sudo is not available" >&2
  exit 1
}

version="${STFW_VERSION:-latest}"
bindir="${STFW_BINDIR:-$(default_bindir)}"
version_set=0
bindir_set=0

while [ "$#" -gt 0 ]; do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    --version)
      shift
      if [ "$#" -eq 0 ]; then
        echo "install.sh: --version requires a value" >&2
        exit 1
      fi
      version="$1"
      version_set=1
      ;;
    --bindir)
      shift
      if [ "$#" -eq 0 ]; then
        echo "install.sh: --bindir requires a value" >&2
        exit 1
      fi
      bindir="$1"
      bindir_set=1
      ;;
    *)
      if [ "${version_set}" -eq 0 ]; then
        version="$1"
        version_set=1
      elif [ "${bindir_set}" -eq 0 ]; then
        bindir="$1"
        bindir_set=1
      else
        echo "install.sh: unexpected argument: $1" >&2
        usage >&2
        exit 1
      fi
      ;;
  esac
  shift
done

need_cmd curl tar uname mktemp install

case "$(uname -s)" in
  Linux) os="linux" ;;
  Darwin) os="darwin" ;;
  *)
    echo "install.sh: unsupported OS: $(uname -s)" >&2
    exit 1
    ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *)
    echo "install.sh: unsupported arch: $(uname -m)" >&2
    exit 1
    ;;
esac

resolved_tag=""
resolved_version=""
normalize_version "${version}"

tarball="${binary_name}_${resolved_version}_${os}_${arch}.tar.gz"
url="https://github.com/${repo}/releases/download/${resolved_tag}/${tarball}"
tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT

echo "install.sh: downloading ${url}"
curl -fsSL "${url}" -o "${tmpdir}/${tarball}"
tar -xzf "${tmpdir}/${tarball}" -C "${tmpdir}"

if [ ! -f "${tmpdir}/${binary_name}" ]; then
  echo "install.sh: ${binary_name} binary not found in ${tarball}" >&2
  exit 1
fi

install_binary "${tmpdir}/${binary_name}" "${bindir}"

echo "install.sh: installed ${binary_name} ${resolved_tag} to ${bindir}/${binary_name}"
"${bindir}/${binary_name}" --version

case ":${PATH}:" in
  *:"${bindir}":*) ;;
  *)
    echo "install.sh: ${bindir} is not in PATH" >&2
    echo "install.sh: add this to your shell profile: export PATH=\"${bindir}:\$PATH\"" >&2
    ;;
esac
