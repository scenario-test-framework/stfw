# daily-balance-sut のイメージ (example 用トイ SUT)。
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 go build -trimpath -o /out/sut ./...

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/sut /sut
EXPOSE 8080
ENTRYPOINT ["/sut"]
