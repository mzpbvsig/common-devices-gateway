FROM golang:1.19.0-alpine3.16

WORKDIR /home/cg

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

COPY ./  ./

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o /main

CMD [ "/main" ]