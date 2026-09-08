FROM golang:1.26.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]