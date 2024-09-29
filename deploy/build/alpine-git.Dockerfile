FROM alpine:latest
LABEL authors="lisa"

RUN apk update
RUN apk add --no-cache git

# docker build -t lisa/alpine:git -f alpine-git.Dockerfile .
