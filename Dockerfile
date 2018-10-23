FROM golang:alpine AS builder
LABEL maintainer="imlonghao <dockerfile@esd.cc>"
WORKDIR /go/src/app
COPY . ./
RUN apk add git && \
    go get -u github.com/golang/dep/cmd/dep && \
    dep ensure -vendor-only && \
    go build -o /app/blog

FROM alpine
WORKDIR /app
COPY --from=builder /app/blog ./
COPY static/ ./static/
COPY views/ ./views/
EXPOSE 8080
ENTRYPOINT [ "/app/blog" ]