FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /todo-server .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /todo-server .
COPY web ./web

ENV TODO_PORT=7540
ENV TODO_DBFILE=/data/scheduler.db

VOLUME ["/data"]
EXPOSE 7540

CMD ["./todo-server"]