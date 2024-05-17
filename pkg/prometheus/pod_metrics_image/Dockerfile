# 使用官方的golang镜像作为构建环境，请注意版本要与mod文件中的版本一致
FROM golang:1.22-alpine as builder

# 设置工作目录
WORKDIR /app


# 将源代码复制到工作目录
COPY . .


# 设置Docker环境中的go代理，后续它从这里下载依赖
ENV GOPROXY=https://goproxy.cn

# 更新依赖项，编译Go应用
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .


# # 采用vendor模式，这样在build时直接使用上述的依赖项；但这要求在外面提前下载好依赖包，不太好
# # 复制预先下载好的依赖项
# COPY /vendor /vendor
# RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -o main .

# 使用scratch作为新的基础镜像
FROM scratch

# 从builder镜像中复制二进制文件到当前镜像
COPY --from=builder /app/main .

# 暴露端口
EXPOSE 2112

# 运行二进制文件
CMD ["./main"]
