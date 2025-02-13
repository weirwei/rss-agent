# 使用官方的 Golang 镜像作为基础镜像
FROM golang:1.22-alpine as builder

ARG APP_NAME
ENV APP_NAME=$APP_NAME

LABEL maintainer="weirwei <nightnessss@163.com>"
# 设置编译目录
WORKDIR $GOPATH/$APP_NAME
COPY go.mod $GOPATH/${APP_NAME}/
COPY go.sum $GOPATH/${APP_NAME}/

# 下载依赖
ENV GOPROXY='https://goproxy.cn,https://goproxy.io,direct'
RUN go mod download

# 编译
COPY . $GOPATH/${APP_NAME}/
RUN go build -o /usr/local/bin/${APP_NAME} main.go

FROM alpine:3.20

ARG APP_NAME
ENV APP_NAME=$APP_NAME

# 设置运行目录
WORKDIR /usr/local/bin/

# 拷贝编译好的程序
COPY --from=builder /usr/local/bin/${APP_NAME} /usr/local/bin/
# 拷贝配置文件
COPY ./conf /usr/local/bin/conf

# 运行程序
CMD ["sh", "-c", "/usr/local/bin/$APP_NAME"]