#!/bin/bash

if [ "$1" == "3" ]; then
    servers=(node2 node3)
elif [ "$1" == "5" ]; then
    servers=(node2 node3 node4 node5)
else
    echo "Usage: $0 [3|5]"
    exit 1
fi

user="root"

for server in "${servers[@]}"; do
    echo "========================================"
    echo "正在连接到服务器: $server"
    echo "========================================"

    ssh "$user"@"$server" <<EOF
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

    echo "创建 data 文件夹"
    mkdir /data/mongo1_data
    mkdir /data/mongo2_data
    mkdir /data/kvrocks_data
    mkdir /data/cassandra_data
    mkdir /data/cassandra_commitlog
    mkdir /data/cassandra_saved_caches
    mkdir /data/couchdb_data

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
