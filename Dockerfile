FROM golang:1.23 AS builder

ARG CGO_ENABLED=0
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o pb .

FROM scratch
WORKDIR /app
COPY --from=builder /app/pb ./pb
ENTRYPOINT ["/app/pb", "serve", "--http=0.0.0.0:8090", "--dev"]