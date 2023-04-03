# 基础镜像
FROM docker.io/library/golang:1.19.3-buster
# 文件系统
ADD ./ /data/lp_market/

# 端口信息
EXPOSE 18080

# 工作目录
WORKDIR /data/lp_market/
# 安装依赖
RUN go build -o lp_market cmd/main.go 
# 环境变量
ENV GO_ENV=production
ENV SERVICE_PORT=18080

CMD ["./lp_market" ]

