#!/bin/bash

# 定义IP地址映射
declare -A node_ips=(
    ["1"]="172.24.58.116"
    ["2"]="172.24.58.114"
    ["3"]="172.24.58.115"
)

PD_IP="172.24.58.116"
WORK_DIR="$HOME/tikv-deploy"
DOWNLOAD_URL="https://download.pingcap.org/tidb-latest-linux-amd64.tar.gz"

# 定义端口检查和进程清理函数
kill_process_on_port() {
    local port=$1
    local pid
    pid=$(lsof -t -i ":$port")
    if [ -n "$pid" ]; then
        echo "Port $port is occupied by process $pid. Terminating this process..."
        kill -9 "$pid"
        sleep 1
    fi
}

# 检查命令行参数
if [ $# -ne 1 ]; then
    echo "Usage: $0 <node_id>"
    echo "node_id must be 1, 2, or 3"
    exit 1
fi

NODE_ID=$1
CURRENT_IP=${node_ips[$NODE_ID]}

if [ -z "$CURRENT_IP" ]; then
    echo "Invalid node_id. Must be 1, 2, or 3"
    exit 1
fi

# 创建工作目录
mkdir -p $WORK_DIR
cd $WORK_DIR

# 检查是否需要下载包
if [ ! -f "tidb-latest-linux-amd64.tar.gz" ]; then
    echo "Downloading TiKV package..."
    wget $DOWNLOAD_URL
    wget http://download.pingcap.org/tidb-latest-linux-amd64.sha256
    
    # 验证包完整性
    echo "Verifying package integrity..."
    sha256sum -c tidb-latest-linux-amd64.sha256
else
    echo "Package already exists, skipping download..."
fi

# 检查是否已解压
if [ ! -d "bin" ]; then
    echo "Extracting package..."
    tar -xzf tidb-latest-linux-amd64.tar.gz
    mv tidb-*-linux-amd64/* .
    rm -rf tidb-*-linux-amd64
fi

# 根据节点ID启动相应的服务
case $NODE_ID in
    "1")
        # 检查PD端口
        kill_process_on_port 2379
        kill_process_on_port 2380

        echo "Starting PD server..."
        nohup ./bin/pd-server --name=pd1 \
            --data-dir=pd1 \
            --client-urls="http://$CURRENT_IP:2379" \
            --peer-urls="http://$CURRENT_IP:2380" \
            --initial-cluster="pd1=http://$CURRENT_IP:2380" \
            --log-file=pd1.log > pd.out 2>&1 &

        # 等待PD启动
        echo "Waiting for PD to start..."
        sleep 5

        # 检查TiKV端口
        kill_process_on_port 20160

        echo "Starting TiKV server..."
        nohup ./bin/tikv-server --pd-endpoints="$PD_IP:2379" \
            --addr="$CURRENT_IP:20160" \
            --data-dir=tikv1 \
            --log-file=tikv1.log > tikv.out 2>&1 &
        ;;
    
    "2"|"3")
        # 检查TiKV端口
        kill_process_on_port 20160

        echo "Starting TiKV server..."
        nohup ./bin/tikv-server --pd-endpoints="$PD_IP:2379" \
            --addr="$CURRENT_IP:20160" \
            --data-dir=tikv$NODE_ID \
            --log-file=tikv$NODE_ID.log > tikv.out 2>&1 &
        ;;
esac

# 等待服务启动
sleep 2

# 验证服务是否正常启动
case $NODE_ID in
    "1")
        if ! lsof -i :2379 >/dev/null; then
            echo "Warning: PD server might have failed to start. Check pd1.log for details."
        else
            echo "PD server started successfully."
        fi
        ;;
esac

if ! lsof -i :20160 >/dev/null; then
    echo "Warning: TiKV server might have failed to start. Check tikv$NODE_ID.log for details."
else
    echo "TiKV server started successfully."
fi

echo "Deployment completed for node $NODE_ID"
echo "You can check the logs in $WORK_DIR"