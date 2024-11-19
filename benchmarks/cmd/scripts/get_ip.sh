#!/bin/bash

# 获取内网 IP
get_internal_ip() {
    # Linux 系统
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # 尝试获取内网 IP，排除 docker 和 localhost 地址
        internal_ip=$(ip -4 addr show | grep inet | grep -v '127.0.0.1' | grep -v 'docker' | grep -v 'br-' | awk '{print $2}' | cut -d'/' -f1 | head -n 1)
    
    # macOS 系统
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        internal_ip=$(ifconfig | grep "inet " | grep -v '127.0.0.1' | awk '{print $2}' | head -n 1)
    fi
    
    echo "$internal_ip"
}

# 获取外网 IP
get_external_ip() {
    # 尝试多个服务来确保可用性
    external_ip=$(curl -s https://api.ipify.org || \
                 curl -s http://ifconfig.me || \
                 curl -s https://ip.sb || \
                 curl -s https://api.ip.sb/ip)
    echo "$external_ip"
}

# 主函数
main() {
    echo "系统类型: $OSTYPE"
    
    internal_ip=$(get_internal_ip)
    echo "内网 IP: $internal_ip"
    
    echo "正在获取外网 IP..."
    external_ip=$(get_external_ip)
    echo "外网 IP: $external_ip"
    
    # 导出为环境变量
    export INTERNAL_IP="$internal_ip"
    export EXTERNAL_IP="$external_ip"
}

# 运行主函数
main