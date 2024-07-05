FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kidsnoter .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root
COPY --from=builder /app/kidsnoter .
COPY --from=builder /app/templates ./templates

LABEL org.opencontainers.image.source=https://github.com/karolistamutis/kidsnoter

EXPOSE 9091

CMD ["./kidsnoter"]