#!/bin/bash

# 这个脚本用于构建../pkg/prometheus/pod_metrics_image之下定义的镜像，请注意使用相对路径
# 后续Pod启动时可以指定使用这个镜像，暴露2112/metrics端口供prometheus监听
# 获取当前脚本的绝对路径；这与在哪里执行这个脚本无关
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# 获取构建目录的绝对路径
BUILD_DIR_NAME="${SCRIPT_DIR}/../pkg/prometheus/pod_metrics_image"


cd ${BUILD_DIR_NAME}
docker build -t my_pod_metrics ${BUILD_DIR_NAME}