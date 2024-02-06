FROM golang:alpine AS builder
LABEL stage=builder
WORKDIR $GOPATH/src/mypackage/myapp/
COPY ./src/* ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod tidy
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /Chika

FROM scratch
COPY --from=builder /Chika /app/Chika
VOLUME /database
ENTRYPOINT ["/app/Chika"]
