#!/bin/bash

# 定义颜色代码
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

verbose=true

log() {
    local color=${2:-$NC}  # 如果没有指定颜色参数，默认使用无颜色
    if [[ "${verbose}" = true ]]; then
        echo -e "${color}$1${NC}"
    fi
}

# Go to the script root directory
cd "$(dirname "$0")" && cd ..

log "$(pwd)"

log "Error: Operation failed" $RED            # 错误消息用红色
log "Warning: Disk space low" $YELLOW        # 警告消息用黄色
log "Info: Process started" $CYAN            # 信息消息用青色
log "Status: Running" $PURPLE                # 状态消息用紫色
log "Regular message"                        # 不指定颜色参数则使用默认颜色