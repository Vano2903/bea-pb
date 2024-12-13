FROM --platform=$BUILDPLATFORM golang:1.23 AS builder

ARG CGO_ENABLED=0
WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download

COPY . .

# RUN go clean
ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target="/root/.cache/go-build" \
    CGO_ENABLED=0 GOOS=linux go build -o pb -installsuffix cgo -ldflags '-w' .

FROM alpine
WORKDIR /app
COPY --from=builder /app/pb ./pb
ENTRYPOINT ["/app/pb", "serve", "--http=0.0.0.0:8090", "--dev"]