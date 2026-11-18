#!/usr/bin/env bash
# verify-pushed-image.sh は、スモークテストに通したローカルのイメージと、レジストリへ push したイメージの
# 中身 (レイヤーの diff_id の並び) が一致することを確かめる。
# リリースでは確認用のビルドと公開用のビルドを分けて行い、2 回目はビルダーのキャッシュを使う。
# キャッシュが外れて作り直されると、確かめていない中身が公開されるため、ここで止める。
#
# 使い方: verify-pushed-image.sh <ローカルのイメージ> <push したイメージ (name@sha256:...)> <platform (例: linux/amd64)>
# 終了コード: 0 一致 / 1 不一致・取得失敗 / 2 使い方の誤り
set -euo pipefail

if [ "$#" -ne 3 ]; then
  echo "usage: $0 <local-image> <pushed-ref> <platform>" >&2
  exit 2
fi
local_image="$1"
pushed_ref="$2"
platform="$3"

local_layers="$(docker image inspect "$local_image" --format '{{json .RootFS.Layers}}' | jq -c '.')"

# push したものが単一のイメージでも、来歴 (provenance) 付きのインデックスでも読めるようにする。
# インデックスの場合、.Image は platform ごとの設定の辞書になる。
pushed_config="$(docker buildx imagetools inspect "$pushed_ref" --format '{{json .Image}}')"
pushed_layers="$(jq -c --arg p "$platform" 'if has("rootfs") then .rootfs.diff_ids else .[$p].rootfs.diff_ids end' <<<"$pushed_config")"

if [ -z "$pushed_layers" ] || [ "$pushed_layers" = "null" ]; then
  echo "FAIL: could not read layers of $pushed_ref for $platform" >&2
  exit 1
fi
if [ "$local_layers" != "$pushed_layers" ]; then
  echo "FAIL: $pushed_ref differs from the smoke-tested $local_image" >&2
  echo "  tested: $local_layers" >&2
  echo "  pushed: $pushed_layers" >&2
  exit 1
fi
echo "ok: $pushed_ref matches the smoke-tested $local_image ($(jq 'length' <<<"$local_layers") layers)"
