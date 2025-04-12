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
ssh-copy-id -i ~/.ssh/id_ed25519.pub root@FILL

ssh -i ~/.ssh/otc_dse root@FILL

scp -i ~/.ssh/otc_dse otc_dse root@FILL:~/.ssh
scp -i ~/.ssh/otc_dse go1.23.3.linux-amd64.tar.gz .tmux.conf root@FILL:~/

scp -i ~/.ssh/otc_dse oreo.tar.gz root@FILL:~/

scp -i ~/.ssh/otc_dse jdk-17_linux-x64.tar.gz epoxy-3m2.jar root@FILL:~/
```

2. 运行以下指令安装并运行 Docker

```shell
ssh -i ~/.ssh/otc_dse root@FILL

sudo yum install -y docker-ce git ripgrep fish tmux

sudo systemctl start docker
```

> Do we need this step?

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

- 3 个节点

```config
Host node1
    HostName 10.206.206.2
    User root
    IdentityFile ~/.ssh/otc_dse
Host node2
    HostName 10.206.206.3
    User root
    IdentityFile ~/.ssh/otc_dse
Host node3
    HostName 10.206.206.4
    User root
    IdentityFile ~/.ssh/otc_dse
```

- 5 个节点

```config
Host node1
    HostName 10.206.206.2
    User root
    IdentityFile ~/.ssh/otc_dse
Host node2
    HostName 10.206.206.3
    User root
    IdentityFile ~/.ssh/otc_dse
Host node3
    HostName 10.206.206.4
    User root
    IdentityFile ~/.ssh/otc_dse
Host node4
    HostName 10.206.206.5
    User root
    IdentityFile ~/.ssh/otc_dse
Host node5
    HostName 10.206.206.6
    User root
    IdentityFile ~/.ssh/otc_dse

```

1. Git Clone

```shell
git clone git@github.com:KKKZOZ/oreo.git --depth=1

git config --global user.email "1206668472@qq.com"

git config --global user.name "KKKZOZ"
```

6. 安装 Golang

```shell
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz

# vim .bashrc
# export PATH=$PATH:/usr/local/go/bin


echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

source .bashrc

go env -w GOPROXY=https://goproxy.cn,direct
```

- 安装 Java

tar -xzf jdk-17_linux-x64.tar.gz

export JAVA_HOME=/root/jdk-17.0.11

7. 执行 `setup.sh` 完成对其他服务器的配置

```shell
cd oreo/benchmarks/cmd/scripts/vm-setup
./setup.sh
```

8. 配置 docker mirror

```shell
sudo mkdir -p /etc/docker

echo '{
  "registry-mirrors": [
    "https://mirror.ccs.tencentyun.com"
  ]
}' | sudo tee /etc/docker/daemon.json

sudo systemctl restart docker
```

8. 在执行 Workload 之前, 需要手动配置各个服务器目前运行的数据库

- Postgres

```shell
docker run -d -p 5432:5432 --rm --name="apiary-postgres" --env POSTGRES_PASSWORD=dbos postgres:latest
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


./ycsb-setup.sh Redis
./ycsb-setup.sh Cassandra

```

- Run

```shell
./ycsb-full.sh -wl A -db MongoDB1,MongoDB2 -v -r
./ycsb-full.sh -wl F -db MongoDB1,MongoDB2 -v -r

./ycsb-full.sh -wl A -db Redis,Cassandra -v -r
./ycsb-full.sh -wl F -db Redis,Cassandra -v -r
```

### Realistic Workloads

```shell
docker volume prune
```

#### IOT

- Setup

```shell
./realistic-setup.sh -wl iot -id 2
./realistic-setup.sh -wl iot -id 3
```

- Run

```shell
./realistic-full.sh -wl iot -v -r
```

#### Social

- Setup

```shell
./realistic-setup.sh -wl social -id 2
./realistic-setup.sh -wl social -id 3
./realistic-setup.sh -wl social -id 4
```

- Run

```shell
./realistic-full.sh -wl social -v -r
```

#### Order

- Setup

```shell
./realistic-setup.sh -wl order -id 2
./realistic-setup.sh -wl order -id 3
./realistic-setup.sh -wl order -id 4
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

./opt-full.sh -wl RRMW -v -r
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

修改 `read-full.sh`，只运行 `readStrategy=p` 的部分，手动修改 `Chain Length`

### Scalability

> 需要换服务器 32 vcpu, 3 台

```shell
# Redis,MongoDB2
# node 2
./ycsb-setup.sh Redis

# node 3
./ycsb-setup.sh MongoDB2


# MongoDB1,Cassandra
# node 2
./ycsb-setup.sh MongoDB1

# node 3
./ycsb-setup.sh Cassandra

```

#### Vertical

> 注意如果直接启动的话, load data 这一步骤会非常慢, 建议先在 scale-full 脚本
> 中的 deploy_remote 中把 `start-exeuctor-docker.sh` 的 `-l` 参数删掉
> 数据加载完成后再加回来, 然后执行 `docker rm -f executor-8001`

> 记得记录 Executor cpu usage

```shell
./scale-full.sh -wl RMW -t 256 -v -r
```

```shell
docker update --cpus=1 executor-8001 && htop
docker update --cpus=2 executor-8001 && htop
docker update --cpus=4 executor-8001 && htop
docker update --cpus=6 executor-8001 && htop
docker update --cpus=8 executor-8001 && htop
docker update --cpus=10 executor-8001 && htop
docker update --cpus=12 executor-8001 && htop
docker update --cpus=14 executor-8001 && htop
docker update --cpus=16 executor-8001 && htop
```

#### Horizontal

```shell
./scale-full.sh -wl RMW -v -r
```

> 记得修改 `BenchmarkConfig_ycsb.yaml` 中的 executor_address_map

```shell
./start-executor-docker.sh -l  -wl ycsb -db MongoDB1,Cassandra -p 8002
./start-executor-docker.sh -l  -wl ycsb -db MongoDB1,Cassandra -p 8003
./start-executor-docker.sh -l  -wl ycsb -db MongoDB1,Cassandra -p 8004
./start-executor-docker.sh -l  -wl ycsb -db MongoDB1,Cassandra -p 8005
./start-executor-docker.sh -l  -wl ycsb -db MongoDB1,Cassandra -p 8006
./start-executor-docker.sh -l  -wl ycsb -db MongoDB1,Cassandra -p 8007
./start-executor-docker.sh -l  -wl ycsb -db MongoDB1,Cassandra -p 8008
```
