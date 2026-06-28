FROM golang:1.26-alpine AS builder
ARG SERVICE
WORKDIR /app
COPY . .
RUN cd services/${SERVICE} && go build -o /app/server ./cmd

FROM alpine:latest
COPY --from=builder /app/server /server
EXPOSE 8080
CMD ["/server"]
