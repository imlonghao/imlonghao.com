FROM golang:alpine AS builder
LABEL maintainer="imlonghao <dockerfile@esd.cc>"
WORKDIR /app
COPY *.go ./
RUN apk add git && \
    go get github.com/jackc/pgx && \
    go get github.com/gin-gonic/gin && \
    go get github.com/gomarkdown/markdown && \
    go get github.com/gorilla/feeds && \
    go build -o /app/blog

FROM alpine
WORKDIR /app
COPY --from=builder /app/blog ./
COPY static/ ./static/
COPY views/ ./views/
EXPOSE 8080
ENTRYPOINT [ "/app/blog" ]