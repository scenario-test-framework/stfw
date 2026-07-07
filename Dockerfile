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
FROM debian:bookworm-slim
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
