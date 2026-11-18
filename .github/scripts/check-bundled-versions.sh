#!/usr/bin/env bash
# check-bundled-versions.sh は、同梱ツールの既定版 (各プラグインの config.yml。Dockerfile もここから
# 版を読む) が各ツールの最新リリースと一致するかを確かめる (AS-BUILT §10.4 の定期チェック)。
# stfw:full の Chrome (Dockerfile の CHROME_VERSION と、公開済みの :full の中身) も Chrome for Testing の Stable と比べる。
#
#   使い方: GH_TOKEN=<token> check-bundled-versions.sh   (リポジトリ直下で実行)
#
# 終了コード: 0 = すべて最新 / 1 = 新しいリリースがある (または取得に失敗)
set -u

# conf は config.yml の値からクオート・インラインコメント・末尾空白/CR を除いて返す (Dockerfile と同じ解釈)
conf() {
  sed -n "s/^ *$2: *//p" "assets/plugins/process/$1/config.yml" \
    | sed -e 's/\r$//' -e 's/[[:space:]]#.*$//' -e "s/[\"']//g" -e 's/[[:space:]]*$//'
}

rc=0
check() {
  local tool="$1" repo="$2" current="$3" latest
  if [ -z "$current" ]; then
    echo "::error::${tool}: version not found in config.yml"
    rc=1
    return
  fi
  if ! latest="$(gh release view -R "$repo" --json tagName -q .tagName)"; then
    echo "::error::${tool}: failed to get the latest release of ${repo}"
    rc=1
    return
  fi
  if [ "$current" = "$latest" ]; then
    echo "${tool}: ${current} (latest)"
  else
    echo "::warning::${tool}: bundled ${current}, latest release ${latest} (https://github.com/${repo}/releases/tag/${latest})"
    rc=1
  fi
}

check k6 grafana/k6 "$(conf invokeRest k6_version)"
check compare-files scenario-test-framework/compare-files "$(conf compare compare_files_version)"
check logfilter scenario-test-framework/logfilter "$(conf collectLog logfilter_version)"

# stfw:full の Chrome (Chrome for Testing の画面なし版)。apk のパッケージではなく Trivy に見えないため、
# 版が Chrome for Testing の Stable と一致するかで取り込み漏れを見る。版は Dockerfile の ARG が正
check_chrome() {
  local current latest json
  current="$(sed -n 's/^ARG CHROME_VERSION=//p' Dockerfile)"
  if [ -z "$current" ]; then
    echo "::error::chrome: CHROME_VERSION not found in Dockerfile"
    rc=1
    return
  fi
  if ! json="$(curl -fsSL https://googlechromelabs.github.io/chrome-for-testing/last-known-good-versions.json)" \
    || ! latest="$(printf '%s' "$json" | jq -er '.channels.Stable.version')"; then
    echo "::error::chrome: failed to get the Stable version of Chrome for Testing"
    rc=1
    return
  fi
  if [ "$current" = "$latest" ]; then
    echo "chrome: ${current} (latest Stable)"
  else
    echo "::warning::chrome: bundled ${current}, latest Stable ${latest} (update CHROME_VERSION and both CHROME_SHA256_* in Dockerfile)"
    rc=1
  fi
  check_published_chrome "$latest"
}

# 公開済みの stfw:full の Chrome も Stable と比べる。master の版を上げても、リリースするまで利用者が
# 取得するイメージの Chrome は古いままで、Trivy にも見えないため、ここで「リリースが要る」を知らせる。
# 中身を実行せずにファイルだけ取り出す (arm64 もエミュレーションなしで読める)。
# PUBLISHED_FULL_IMAGE を空にすると確認しない (手元で docker が無いときなど)
check_published_chrome() {
  local latest="$1" image="${PUBLISHED_FULL_IMAGE-ghcr.io/scenario-test-framework/stfw:full}" platform cid ver
  [ -n "$image" ] || return 0
  for platform in linux/amd64 linux/arm64; do
    ver=""
    if docker pull -q --platform "$platform" "$image" >/dev/null \
      && cid="$(docker create --platform "$platform" "$image")"; then
      ver="$(docker cp "${cid}:/opt/chrome/VERSION" - 2>/dev/null | tar -xO 2>/dev/null | tr -d '[:space:]')"
      docker rm "$cid" >/dev/null
    else
      echo "::error::chrome: failed to read ${image} (${platform})"
      rc=1
      continue
    fi
    if [ "$ver" = "$latest" ]; then
      echo "chrome in ${image} (${platform}): ${ver} (latest Stable)"
    else
      echo "::warning::chrome in ${image} (${platform}): ${ver:-not bundled}, latest Stable ${latest} (release a new version)"
      rc=1
    fi
  done
}
check_chrome
exit "$rc"
