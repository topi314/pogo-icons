FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build -o pogo-icons github.com/topi314/pogo-icons

FROM alpine

COPY --from=build /build/pogo-icons /bin/pogo-icons

ENTRYPOINT ["/bin/pogo-icons"]

CMD ["-config", "/var/lib/config.toml"]
