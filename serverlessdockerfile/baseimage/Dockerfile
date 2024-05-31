FROM python:3.8

# 设置工作目录
WORKDIR /app

# 复制应用程序代码到容器
COPY . /app

# 使用 Tsinghua 大学镜像来配置 pip
# RUN pip3 config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple

# 更新包列表并安装必要的依赖
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    libmagickwand-dev \
    imagemagick && \
    rm -rf /var/lib/apt/lists/*

# 安装 Python 依赖
RUN pip install -r requirements.txt

# 暴露端口 10000
EXPOSE 10000

# 设置运行应用程序的命令
CMD ["python", "baseserver.py"]
