FROM golang:1.27 AS build
ARG VERSION=1.0.0-dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath \
      -ldflags "-s -w -X github.com/scenario-test-framework/stfw/internal/presentation/cli.Version=${VERSION}" \
      -o /out/stfw ./cmd/stfw

# distroless にしない理由: プロセスプラグイン契約が「任意言語のユーザースクリプト実行」
# のため、最低限 bash + ssh クライアントが必要
FROM debian:bookworm-slim AS runtime
RUN apt-get update \
 && apt-get install -y --no-install-recommends bash curl openssh-client ca-certificates \
 && rm -rf /var/lib/apt/lists/* \
 && useradd --uid 1000 --create-home stfw \
 # reports named volume の初回初期化でこの所有権が引き継がれる (uid 1000 で書けるようにする)
 && mkdir -p /work/.stfw/reports \
 && chown -R stfw:stfw /work
COPY --from=build /out/stfw /usr/local/bin/stfw
USER stfw
WORKDIR /work
ENTRYPOINT ["stfw"]

# stfw:full — 全組込みプラグインのランタイム同梱版。
#   sshpass                   : collectFile / collectLog / sshExec / scpPut (ssh/scp 認証)
#   default-mysql-client      : export/import/clearMysql (MariaDB ベースの mysql クライアント)
#   postgresql-client         : export/import/clearPostgres (psql)
#   redis-tools               : export/import/clearRedis (redis-cli)
#   chromium + fonts-noto-cjk : invokeWeb (k6 ブラウザモード)
#   compare-files / k6 / logfilter : compare / invokeRest・invokeWeb / collectLog の外部ツール
#                                    (/opt/stfw/bundled。stfw plugin install なしで実行できる)
# 通常版のビルドは --target runtime を明示する (無指定の docker build は最終ステージ = full)。
FROM runtime AS full
USER root
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      sshpass \
      default-mysql-client \
      postgresql-client \
      redis-tools \
      chromium \
      fonts-noto-cjk \
 && rm -rf /var/lib/apt/lists/*
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
# k6 browser の Chromium 解決先 (K6_BROWSER_EXECUTABLE_PATH)。
# コンテナ内は seccomp で user namespace を作れず Chromium sandbox が起動しないため
# no-sandbox を既定にする (K6_BROWSER_ARGS はカンマ区切り・`--` なしの k6 形式)。
ENV K6_BROWSER_EXECUTABLE_PATH=/usr/bin/chromium \
    K6_BROWSER_ARGS=no-sandbox
USER stfw
