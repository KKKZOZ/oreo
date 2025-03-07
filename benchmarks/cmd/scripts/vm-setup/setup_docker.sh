#!/bin/bash

# 服务器列表
servers=("10.206.206.3" "10.206.206.4" "10.206.206.5" "10.206.206.6")
# servers=("10.206.206.3" "10.206.206.4")

# 用户名
user="root"

# 循环遍历服务器列表
for server in "${servers[@]}"; do
    echo "========================================"
    echo "正在连接到服务器: $server"
    echo "========================================"

    # 使用ssh连接到服务器并执行命令
    ssh -o StrictHostKeyChecking=no "$user"@"$server" <<EOF
    echo "执行命令: sudo yum install -y docker-ce"
    sudo yum install -y docker-ce
    if [ $? -ne 0 ]; then
        echo "命令执行失败: sudo yum install -y docker-ce"
        exit 1
    fi

    echo "执行命令: sudo systemctl start docker"
    sudo systemctl start docker
    if [ $? -ne 0 ]; then
        echo "命令执行失败: sudo systemctl start docker"
        exit 1
    fi

    sudo yum install -y fish htop

    # 创建 /etc/docker 目录
    echo "执行命令: mkdir -p /etc/docker"
    sudo mkdir -p /etc/docker
    if [ $? -ne 0 ]; then
        echo "命令执行失败: mkdir -p /etc/docker"
        exit 1
    fi

    # 写入 daemon.json 文件
    echo "执行命令: 写入 /etc/docker/daemon.json"
    echo '{
  "registry-mirrors": [
    "https://mirror.ccs.tencentyun.com"
  ]
}' | sudo tee /etc/docker/daemon.json
    if [ $? -ne 0 ]; then
        echo "命令执行失败: 写入 /etc/docker/daemon.json"
        exit 1
    fi
    
    # 重启 docker
    echo "执行命令: sudo systemctl restart docker"
    sudo systemctl restart docker
    if [ $? -ne 0 ]; then
        echo "命令执行失败: sudo systemctl restart docker"
        exit 1
    fi

    echo "在 $server 上执行命令完成"
    exit
EOF

    if [ $? -ne 0 ]; then
        echo "连接到服务器 $server 或执行命令失败"
    else
        echo "在服务器 $server 上执行命令成功"
    fi
done

echo "所有服务器操作完成"
