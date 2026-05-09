FROM golang:1.24-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct
ENV GOTELEMETRY=off
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /server .
COPY config/config.yaml ./config/

EXPOSE 8080

CMD ["./server"]
