FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /app/server ./cmd

FROM alpine:latest
COPY --from=builder /app/server /server
EXPOSE 8080
CMD ["/server"]
