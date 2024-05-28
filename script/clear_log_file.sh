#!/bin/bash
# 这个脚本会读取constant.go文件中对于LOG_FILE_PATH的赋值（它是一个相对于当前用户主目录的路径），然后如果存在的话、删掉它；注意，这需要正确指定constant.go相对本文件的路径关系
# 请勿sudo这个文件，因为sudo会改变$HOME的值，导致找不到正确的日志文件路径

# 获取当前脚本的绝对路径；这与在哪里执行这个脚本无关
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# 指定相对路径../pkg/constant/constant.go后，从 constant.go 文件中提取 LOG_FILE_PATH 的值；应该假定LOG_FILE_PATH = ...的语句在constant.go文件中唯一出现！
LOG_FILE_PATH=$(grep -oP 'LOG_FILE_PATH_DEFAULT = "\K[^"]+' $SCRIPT_DIR/../pkg/constant/logger_constant.go)

# 拼接用户主目录的路径
LOG_FILE="$HOME/$LOG_FILE_PATH"

# 删除日志文件
if [ -f "$LOG_FILE" ]; then
    rm "$LOG_FILE"
    echo "Log file $LOG_FILE has been deleted."
else
    echo "Log file $LOG_FILE does not exist, no need to delete."
fi
