# syntax=docker/dockerfile:1

FROM golang:1.23 AS build-stage

WORKDIR /app

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o ./go-disbot

FROM alpine:latest AS release-stage
# FROM gcr.io/distroless/base-debian11 AS release-stage

WORKDIR /app

COPY --from=build-stage /app/go-disbot /app/.env.vault /app/challenges.json ./

RUN chmod 644 /app/challenges.json

ENTRYPOINT [ "/app/go-disbot" ]

