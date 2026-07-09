FROM golang:1.26 AS build
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
#   chromium + fonts-noto-cjk : invokeWeb (k6 ブラウザモード。k6 本体は stfw plugin install が取得)
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
# k6 browser の Chromium 解決先 (K6_BROWSER_EXECUTABLE_PATH)。
# コンテナ内は seccomp で user namespace を作れず Chromium sandbox が起動しないため
# no-sandbox を既定にする (K6_BROWSER_ARGS はカンマ区切り・`--` なしの k6 形式)。
ENV K6_BROWSER_EXECUTABLE_PATH=/usr/bin/chromium \
    K6_BROWSER_ARGS=no-sandbox
USER stfw
