## Script

```bash
#!/bin/bash

# 检查参数
if [ $# -ne 1 ]; then
    echo "Usage: $0 <public_ip>"
    exit 1
fi

PUBLIC_IP=$1
PASSWORD="Qwert12345??"

# 确保 .ssh 目录存在
mkdir -p ~/.ssh

# 1. 配置免密登录到公网 IP
echo "Configuring passwordless login to public IP..."
sshpass -p "${PASSWORD}" ssh-copy-id -o StrictHostKeyChecking=no -i ~/.ssh/id_ed25519.pub root@${PUBLIC_IP}
scp ~/.ssh/id_ed25519 root@${PUBLIC_IP}:~/.ssh/
scp ~/go1.23.3.linux-amd64.tar.gz ~/.tmux.conf root@${PUBLIC_IP}:~/

# 2. 登录并安装必要软件
echo "Installing required software..."
ssh root@${PUBLIC_IP} << 'EOF'
yum install -y docker-ce git ripgrep fish tmux sshpass
systemctl start docker

# 3. 配置内网机器免密登录
ssh-keygen -t rsa -N "" -f ~/.ssh/id_rsa
for ip in "10.206.206.3" "10.206.206.4"; do
    sshpass -p "${PASSWORD}" ssh-copy-id -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa.pub root@$ip
done

# 4. 创建 SSH 配置文件
cat > ~/.ssh/config << 'EOL'
Host node1
    HostName 10.206.206.2
    User root
    IdentityFile ~/.ssh/id_rsa
Host node2
    HostName 10.206.206.3
    User root
    IdentityFile ~/.ssh/id_rsa
Host node3
    HostName 10.206.206.4
    User root
    IdentityFile ~/.ssh/id_rsa
EOL

# 5. Git 配置和克隆仓库
git config --global user.email "1206668472@qq.com"
git config --global user.name "KKKZOZ"
git clone git@github.com:KKKZOZ/oreo.git --depth=1

# 6. 安装 Golang
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go env -w GOPROXY=https://goproxy.cn,direct
EOF

echo "Setup completed successfully!"


```

## Manual

1. 使用公网 IP 登录 Node1

```shell
ssh root@FILL

scp id_ed25519 root@FILL:~/.ssh
scp go1.23.3.linux-amd64.tar.gz .tmux.conf root@FILL:~/

# scp oreo.tar.gz root@FILL:~/

scp jdk-17_linux-x64.tar.gz epoxy.jar root@FILL:~/
```

2. 运行以下指令安装并运行 Docker

```shell
ssh root@FILL

sudo yum install -y docker-ce git ripgrep fish tmux iproute-tc

sudo mkdir -p /etc/docker

echo '{
  "registry-mirrors": [
    "https://mirror.ccs.tencentyun.com"
  ]
}' | sudo tee /etc/docker/daemon.json

sudo systemctl start docker
```

> 如果服务器已经配置了秘钥登录, 则下面的第三步不需要执行

3. 配置其他几台服务器的免密登录

```shell
ssh-keygen -t rsa

ssh-copy-id -i ~/.ssh/id_rsa.pub root@10.206.206.3
ssh-copy-id -i ~/.ssh/id_rsa.pub root@10.206.206.4
ssh-copy-id -i ~/.ssh/id_rsa.pub root@10.206.206.5
ssh-copy-id -i ~/.ssh/id_rsa.pub root@10.206.206.6 
```

4. 设置 `~/.ssh/config`

> vim ~/.ssh/config

```config
Host node1
    HostName 10.206.206.2
    User root
Host node2
    HostName 10.206.206.3
    User root
Host node3
    HostName 10.206.206.4
    User root
Host node4
    HostName 10.206.206.5
    User root
Host node5
    HostName 10.206.206.6
    User root

```

1. Git Clone

```shell
git clone git@github.com:KKKZOZ/oreo.git --depth=1

git config --global user.email "1206668472@qq.com" && git config --global user.name "KKKZOZ"
```

6. 安装 Golang 和 Java

```shell
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz

tar -xzf jdk-17_linux-x64.tar.gz

# vim .bashrc
# export PATH=$PATH:/usr/local/go/bin

echo 'export JAVA_HOME=/root/jdk-17.0.11' >> ~/.bashrc
echo 'export PATH=$PATH:/usr/local/go/bin:$JAVA_HOME/bin' >> ~/.bashrc

source .bashrc && go env -w GOPROXY=https://goproxy.cn,direct
```

1. 执行 `setup.sh` 完成对其他服务器的配置

```shell
cd oreo/benchmarks/cmd/scripts/vm-setup
./setup.sh

../setup/update-essentials.sh 3
```

1. 同步数据

```shell
rsync -avP root@FILL:~/oreo/benchmarks/cmd/data/ ~/Projects/oreo/benchmarks/cmd/data/


rsync -avP root@FILL:~/oreo ~/Projects/oreo2
```

## Run Evaluation

### YCSB

- Setup

```shell
# Node 2
./ycsb-setup.sh MongoDB1
# Node 3
./ycsb-setup.sh MongoDB2

# Node 2
./ycsb-setup.sh Redis
# Node 3
./ycsb-setup.sh Cassandra

```

- Run

```shell
./ycsb-full.sh -wl A -db MongoDB1,MongoDB2 -v -r
./ycsb-full.sh -wl F -db MongoDB1,MongoDB2 -v -r

./ycsb-full.sh -wl A -db Redis,Cassandra -v -r
./ycsb-full.sh -wl F -db Redis,Cassandra -v -r
```

#### Epoxy

```shell
# Node 2
docker run -d -p 5432:5432 --rm --name="apiary-postgres" --env POSTGRES_PASSWORD=dbos postgres:latest

java -jar epoxy
```

- Load Data

```shell
POST /load?recordCount=1000000&threadNum=50&mongoIdx=1 HTTP/1.1
Content-Length: 0
Host: 119.45.11.232:7081
User-Agent: HTTPie

POST /load?recordCount=1000000&threadNum=50&mongoIdx=2 HTTP/1.1
Content-Length: 0
Host: 119.45.11.232:7081
User-Agent: HTTPie

```

- Run

```shell
GET /workloada?operationCount=60000&threadNum=8 HTTP/1.1
Host: 119.45.11.232:7081
User-Agent: HTTPie

GET /workloada?operationCount=60000&threadNum=16 HTTP/1.1
Host: 119.45.11.232:7081
User-Agent: HTTPie

GET /workloadf?operationCount=60000&threadNum=8 HTTP/1.1
Host: 119.45.11.232:7081
User-Agent: HTTPie

GET /workloadf?operationCount=60000&threadNum=16 HTTP/1.1
Host: 119.45.11.232:7081
User-Agent: HTTPie

```

### Realistic Workloads

```shell
docker volume prune
```

#### IOT

- Setup

```shell
# Node 2
./realistic-setup.sh -wl iot -id 2
# Node 3
./realistic-setup.sh -wl iot -id 3
```

- Run

```shell
./realistic-full.sh -wl iot -v -r
```

#### Social

- Setup

```shell
# Node 2
./realistic-setup.sh -wl social -id 2
# Node 3
./realistic-setup.sh -wl social -id 3
# Node 4
./realistic-setup.sh -wl social -id 4
```

- Run

```shell
./realistic-full.sh -wl social -v -r
```

#### Order

- Setup

```shell
# Node 2
./realistic-setup.sh -wl order -id 2
# Node 3
./realistic-setup.sh -wl order -id 3
# Node 4
./realistic-setup.sh -wl order -id 4
# Node 5
./realistic-setup.sh -wl order -id 5
```

- Run

```shell
./realistic-full.sh -wl order -v -r
```

### Optimization

- Setup

```shell
./opt-setup.sh -id 2
./opt-setup.sh -id 3
```

- Run

> 注意查看配置文件中的 TxnOperationGroup, zipfian 等参数

- TxnOperationGroup = 8
- zipfian_constant  = 0.8

```shell
./opt-full.sh -wl RMW -v -r

./opt-full.sh -wl RW -v -r
```

### Read Strategy

- Setup

```shell
./read-setup.sh -id 2
./read-setup.sh -id 3
```

- Run

> 注意查看配置文件中的 TxnOperationGroup, zipfian 等参数
> 因为在 OPT 部分修改过

- TxnOperationGroup = 6
- zipfian_constant  = 0.9

```shell
./read-full.sh -wl RMW -v -r

./read-full.sh -wl RRMW -v -r
```

### Scalability && FT

#### FT

> Deploy HAProxy on node2

```shell
sudo yum install -y haproxy
sudo cp haproxy.cfg /etc/haproxy/haproxy.cfg

sudo systemctl start haproxy

sudo systemctl status haproxy

sudo systemctl restart haproxy

sudo systemctl stop haproxy

# Optional
haproxy -f ./haproxy.cfg
```

- Setup Databases

```shell
# node2
./ycsb-setup.sh Redis
# node3
./ycsb-setup.sh MongoDB1
# node4
./ycsb-setup.sh Cassandra
```

- Load Data

```shell
./ft-full.sh -r
```

- Run

```shell
./ft-full.sh -r -l

# When pressing enter, start another shell, run
./ft-process.sh
```

#### Horizontal Scalability

```shell
# MongoDB1,Cassandra
# node 2
./ycsb-setup.sh MongoDB1

# node 3
./ycsb-setup.sh Cassandra

```

- Load

```shell
./scale-full.sh -wl RMW -v -r -n 6
```

- Run

```shell
./scale-full.sh -wl RMW -v -r -n 1
./scale-full.sh -wl RMW -v -r -n 2
./scale-full.sh -wl RMW -v -r -n 3
./scale-full.sh -wl RMW -v -r -n 4
./scale-full.sh -wl RMW -v -r -n 5
./scale-full.sh -wl RMW -v -r -n 6
./scale-full.sh -wl RMW -v -r -n 7
./scale-full.sh -wl RMW -v -r -n 8
```
