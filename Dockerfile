# 使用基础镜像
FROM 192.168.99.14:8101/golang:1.22.2-bullseye
# 设置维护者信息
LABEL maintainer="zhaoyi <tom1120@126.com>"

# 安装 ffmpeg
RUN export http_proxy=http://192.168.96.18:7890 && export https_proxy=http://192.168.96.18:7890 && \
    apt-get update && \
    apt-get install -y ffmpeg && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# 设置工作目录
WORKDIR /data
# 复制项目文件
# COPY build /data
# COPY code /data
# COPY main.go /data
# COPY go.mod /data
# COPY go.sum /data
COPY . /data
# 设置 GOPROXY 环境变量
ENV GOPROXY=https://goproxy.io
# 下载依赖并编译应用
RUN go mod tidy && go build -mod=readonly -o /data/build/tank
# 清理下载的模块缓存
RUN go clean -modcache && export http_proxy= && export https_proxy=

# 暴露端口
EXPOSE 6010
# 设置卷
VOLUME /data/build/matter
# 设置启动命令
ENTRYPOINT ["/data/build/tank"]