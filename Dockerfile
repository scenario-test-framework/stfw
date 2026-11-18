FROM golang:1.27 AS build
ARG VERSION=1.0.0-dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath \
      -ldflags "-s -w -X github.com/scenario-test-framework/stfw/internal/presentation/cli.Version=${VERSION}" \
      -o /out/stfw ./cmd/stfw

# 実行イメージのベースは Wolfi (Chainguard のコンテナ専用ディストリビューション。glibc・apk)。
# Debian より部品が少なく、修正版への追従が早いため、直しようのない脆弱性を抱えにくい (AS-BUILT §10.2)。
# 無料で取れるのは latest だけなので digest で固定し、Dependabot (docker) が digest を更新する。
# ビルド時に apk upgrade して、digest が古くてもパッケージは最新にする。
# distroless にしない理由: プロセスプラグイン契約が「任意言語のユーザースクリプト実行」
# のため、最低限 bash + ssh クライアントが必要
FROM cgr.dev/chainguard/wolfi-base:latest@sha256:580c9ae79c0229cc381475fbfbed0d91fab5d6606e57cd78732f9836de109e90 AS base

# Google 公式の Chrome (Chrome for Testing の画面なし版 chrome-headless-shell)。invokeWeb (k6 ブラウザモード) 用。
# Wolfi の chromium パッケージは更新が止まっているため使わない。apk のパッケージではないので Trivy には見えない。
# 代わりに版をここで固定し、定期チェック (check-bundled-versions.sh) が Chrome for Testing の Stable と比べる。
# 配布物に公式のチェックサムが無いため、版を上げるときに取得して記録し、ビルド時に照合する。
# 版を上げる手順: 版と両 arch の sha256 の 3 つを同時に変える (AS-BUILT §10.2)。
FROM base AS chrome
ARG TARGETARCH
ARG CHROME_VERSION=158.0.8127.36
ARG CHROME_SHA256_AMD64=2d3a23017323c54fd457a678e4351f895c4076d8050e39d5aa6247f93448b73a
ARG CHROME_SHA256_ARM64=176ee5c55486c22de43128871189061365ea28e881fc96016fd18e7cb59437c8
RUN set -eu; \
    apk add --no-cache curl unzip; \
    case "${TARGETARCH}" in \
      amd64) plat=linux64; sum="${CHROME_SHA256_AMD64}" ;; \
      arm64) plat=linux-arm64; sum="${CHROME_SHA256_ARM64}" ;; \
      *) echo "unsupported arch: ${TARGETARCH}" >&2; exit 1 ;; \
    esac; \
    curl -fsSL -o /tmp/chrome.zip \
      "https://storage.googleapis.com/chrome-for-testing-public/${CHROME_VERSION}/${plat}/chrome-headless-shell-${plat}.zip"; \
    echo "${sum}  /tmp/chrome.zip" | sha256sum -c -; \
    unzip -q /tmp/chrome.zip -d /tmp/chrome; \
    mv "/tmp/chrome/chrome-headless-shell-${plat}" /opt/chrome; \
    echo "${CHROME_VERSION}" > /opt/chrome/VERSION

FROM base AS runtime
# GNU 版の基本コマンド (coreutils など) も入れる: wolfi-base の既定は BusyBox の簡易版で、
# 利用者のスクリプトが GNU 版のオプション (Debian ベースの頃と同じ) を前提にしていても動くようにする
RUN apk upgrade --no-cache \
 && apk add --no-cache bash curl openssh-client ca-certificates \
      coreutils findutils grep sed gawk gnutar gzip diffutils \
 && addgroup -g 1000 stfw \
 && adduser -D -u 1000 -G stfw -s /bin/bash stfw \
 # reports named volume の初回初期化でこの所有権が引き継がれる (uid 1000 で書けるようにする)
 && mkdir -p /work/.stfw/reports \
 && chown -R stfw:stfw /work
COPY --from=build /out/stfw /usr/local/bin/stfw
USER stfw
WORKDIR /work
ENTRYPOINT ["stfw"]

# stfw:full — 全組込みプラグインのランタイム同梱版。
#   sshpass                       : collectFile / collectLog / sshExec / scpPut (ssh/scp 認証)
#   mariadb-13.1-client           : export/import/clearMysql (mysql。MariaDB 版クライアント)
#     + mariadb-connector-c       :   MySQL 8 以降の認証方式 (caching_sha2_password) の部品
#   postgresql-18-client          : export/import/clearPostgres (psql)
#   valkey-9.1-cli                : export/import/clearRedis (redis-cli。Valkey は Redis 互換の後継)
#   Chrome + 依存ライブラリ + font-noto-cjk : invokeWeb (k6 ブラウザモード。Chrome は上の chrome ステージ)
#   compare-files / k6 / logfilter : compare / invokeRest・invokeWeb / collectLog の外部ツール
#                                    (/opt/stfw/bundled。stfw plugin install なしで実行できる)
# 通常版のビルドは --target runtime を明示する (無指定の docker build は最終ステージ = full)。
FROM runtime AS full
USER root
# - Wolfi には mysql / redis-cli の名前が無いため、プラグインが呼ぶ名前を用意する。
#   mysql は名前の非推奨警告を出さないよう、symlink ではなく mariadb を呼ぶラッパーにする
# - MariaDB 11.4 以降のクライアントは既定でサーバー証明書を検証し、自己署名の MySQL 8 に繋がらない。
#   Debian ベースの頃 (MariaDB 10.11 クライアント) と同じ「検証しない」を既定に戻す。
#   認証の部品は mariadb-connector-c の置き場所を指す
# - Chrome の依存は共有ライブラリ名 (so:) で指定する (Chrome の版上げで足りなくなったら、ビルド後の
#   スモークテストで分かる。amd64 版は arm64 版より依存が 1 つ多い: libatk-bridge)
RUN apk add --no-cache \
      sshpass \
      mariadb-13.1-client mariadb-connector-c \
      postgresql-18-client \
      valkey-9.1-cli \
      font-noto-cjk \
      so:libX11.so.6 so:libXcomposite.so.1 so:libXdamage.so.1 so:libXext.so.6 so:libXfixes.so.3 \
      so:libXrandr.so.2 so:libasound.so.2 so:libatk-1.0.so.0 so:libatk-bridge-2.0.so.0 so:libatspi.so.0 so:libdbus-1.so.3 \
      so:libexpat.so.1 so:libgbm.so.1 so:libgio-2.0.so.0 so:libglib-2.0.so.0 so:libgobject-2.0.so.0 \
      so:libnspr4.so so:libnss3.so so:libnssutil3.so so:libudev.so.1 so:libxcb.so.1 so:libxkbcommon.so.0 \
 && ln -s /usr/bin/valkey-cli /usr/local/bin/redis-cli \
 && printf '#!/bin/sh\nexec /usr/bin/mariadb "$@"\n' > /usr/local/bin/mysql \
 && chmod 755 /usr/local/bin/mysql \
 && printf '[client]\nssl-verify-server-cert=0\nplugin-dir=/usr/lib/mariadb/plugin\n' > /etc/my.cnf
COPY --from=chrome /opt/chrome /opt/chrome
# 組込みプラグインの外部ツールの同梱。各プラグインの bin/install/install をビルド時に
# stfw_plugin_cache_dir=/opt/stfw/bundled で実行し、{tool}/{os_arch}/{bin} へ配置する。
# 版 (と logfilter の arches) は各プラグインの既定設定 config.yml を正とし、{tool}/VERSION に
# 記録する。プラグインは永続キャッシュが無く、設定版と VERSION が一致するときだけ同梱版を使う。
# k6 は invokeRest / invokeWeb で共有するため、両者の k6_version が食い違えばビルドを失敗させる。
# プラグイン資産は bind mount で読む (COPY すると層がイメージに残るため)。
# conf は config.yml の値からクオート・インラインコメント・末尾空白/CR を除いて返す。
RUN --mount=type=bind,source=assets/plugins/process,target=/tmp/plugins \
    set -eu; \
    p=/tmp/plugins; b=/opt/stfw/bundled; \
    conf() { sed -n "s/^ *$2: *//p" "${p}/$1/config.yml" \
      | sed -e 's/\r$//' -e 's/[[:space:]]#.*$//' -e "s/[\"']//g" -e 's/[[:space:]]*$//'; }; \
    v="$(conf compare compare_files_version)"; test -n "${v}"; \
    stfw_process_compare_compare_files_version="${v}" stfw_plugin_cache_dir="${b}" \
      bash "${p}/compare/bin/install/install"; \
    echo "${v}" > "${b}/compare-files/VERSION"; \
    v="$(conf invokeRest k6_version)"; test -n "${v}"; \
    test "$(conf invokeWeb k6_version)" = "${v}"; \
    stfw_process_invokeRest_k6_version="${v}" stfw_plugin_cache_dir="${b}" \
      bash "${p}/invokeRest/bin/install/install"; \
    echo "${v}" > "${b}/k6/VERSION"; \
    v="$(conf collectLog logfilter_version)"; test -n "${v}"; \
    arches="$(conf collectLog logfilter_arches)"; test -n "${arches}"; \
    stfw_process_collectLog_logfilter_version="${v}" \
    stfw_process_collectLog_logfilter_arches="${arches}" stfw_plugin_cache_dir="${b}" \
      bash "${p}/collectLog/bin/install/install"; \
    echo "${v}" > "${b}/logfilter/VERSION"
ENV stfw_plugin_bundled_dir=/opt/stfw/bundled
# k6 browser の Chrome 解決先 (K6_BROWSER_EXECUTABLE_PATH)。
# コンテナ内は seccomp で user namespace を作れず Chrome の sandbox が起動しないため
# no-sandbox を既定にする (K6_BROWSER_ARGS はカンマ区切り・`--` なしの k6 形式)。
ENV K6_BROWSER_EXECUTABLE_PATH=/opt/chrome/chrome-headless-shell \
    K6_BROWSER_ARGS=no-sandbox
USER stfw
