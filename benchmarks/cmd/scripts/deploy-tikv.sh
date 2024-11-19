#!/bin/bash

# 设置错误时退出
# set -e
IP=$(curl -s https://api.ipify.org || \
                 curl -s http://ifconfig.me || \
                 curl -s https://ip.sb || \
                 curl -s https://api.ip.sb/ip)

echo "Current IP: $IP"

# 定义进程终止函数
kill_process_on_port() {
    local port=$1
    local pid
    pid=$(lsof -t -i ":$port")
    if [ -n "$pid" ]; then
        echo "Port $port is occupied by process $pid. Terminating this process..."
        kill -9 "$pid"
    fi
}

# 创建部署目录
DEPLOY_DIR="tikv-deploy"
mkdir -p $DEPLOY_DIR
cd $DEPLOY_DIR

# 检查是否需要下载包
if [ ! -f "tidb-latest-linux-amd64.tar.gz" ]; then
    echo "Downloading TiKV package..."
    wget https://download.pingcap.org/tidb-latest-linux-amd64.tar.gz
    wget http://download.pingcap.org/tidb-latest-linux-amd64.sha256
    
    # 验证包完整性
    echo "Verifying package integrity..."
    sha256sum -c tidb-latest-linux-amd64.sha256
else
    echo "Package already exists, skipping download..."
fi

# 检查是否已解压
if [ ! -d "tidb-latest-linux-amd64" ]; then
    echo "Extracting package..."
    tar -xzf tidb-latest-linux-amd64.tar.gz
fi

cd tidb-latest-linux-amd64

# 创建必要的数据目录
mkdir -p pd1 tikv1

# 清理可能占用的端口
echo "Checking and clearing ports if necessary..."
kill_process_on_port 2379  # PD client URL port
kill_process_on_port 2380  # PD peer URL port
kill_process_on_port 20160 # TiKV port

# 清理数据文件
echo "Cleaning up data files..."
rm -rf pd1/*
rm -rf tikv1/*
rm -f *.log


# 启动 PD
echo "Starting PD server..."
./bin/pd-server --name=pd1 \
    --data-dir=pd1 \
    --client-urls="http://0.0.0.0:2379" \
    --advertise-client-urls="http://$IP:2379" \
    --peer-urls="http://0.0.0.0:2380" \
    --advertise-peer-urls="http://$IP:2380" \
    --initial-cluster="pd1=http://$IP:2380" \
    --log-file=pd1.log &

# 等待 PD 启动
echo "Waiting for PD to start..."
sleep 1

# 启动 TiKV
echo "Starting TiKV server..."
./bin/tikv-server --pd-endpoints="127.0.0.1:2379" \
    --addr="0.0.0.0:20160" \
    --advertise-addr="$IP:20160" \
    --data-dir=tikv1 \
    --log-file=tikv1.log &

# 等待 TiKV 启动
echo "Waiting for TiKV to start..."
sleep 3

# 验证部署
echo "Verifying deployment..."
curl http://localhost:2379/pd/api/v1/stores

echo "Deployment completed. Servers are running in background."

# 显示运行状态
echo -e "\nCurrent running processes:"
ps aux | grep -E 'pd-server|tikv-server' | grep -v grep
