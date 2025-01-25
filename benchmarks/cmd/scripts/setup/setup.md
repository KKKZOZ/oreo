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
ssh-copy-id -i ~/.ssh/id_ed25519.pub root@49.233.36.106
scp id_ed25519 root@49.233.36.106:~/.ssh
scp go1.23.3.linux-amd64.tar.gz .tmux.conf root@49.233.36.106:~/

scp oreo.tar.gz root@49.233.36.106:~/

scp jdk-17_linux-x64.tar.gz website-0.0.1-SNAPSHOT.jar root@49.233.36.106:~/
```

2. 运行以下指令安装并运行 Docker

```shell
ssh root@49.233.36.106

sudo yum install -y docker-ce git ripgrep fish tmux

sudo systemctl start docker
```

3. 配置其他几台服务器的免密登录

```shell
ssh-keygen -t rsa

ssh-copy-id -i ~/.ssh/id_rsa.pub root@10.206.206.3
ssh-copy-id -i ~/.ssh/id_rsa.pub root@10.206.206.4
ssh-copy-id -i ~/.ssh/id_rsa.pub root@10.206.206.5
ssh-copy-id -i ~/.ssh/id_rsa.pub root@10.206.206.6
```

4. 设置 `~/.ssh/config`

- 3 个节点

```config
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
```

- 5 个节点

```config
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
Host node4
    HostName 10.206.206.5
    User root
    IdentityFile ~/.ssh/id_rsa
Host node5
    HostName 10.206.206.6
    User root
    IdentityFile ~/.ssh/id_rsa

```

5. Git Clone

```shell
git clone git@github.com:KKKZOZ/oreo.git --depth=1

git config --global user.email "1206668472@qq.com"

git config --global user.name "KKKZOZ"
```

6. 安装 Golang

```shell
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz

vim .bashrc

export PATH=$PATH:/usr/local/go/bin

source .bashrc

go env -w GOPROXY=https://goproxy.cn,direct
```

- 安装 Java

tar -xzvf jdk-17_linux-x64.tar.gz

export JAVA_HOME=/root/jdk-17.0.11

7. 执行 `setup.sh` 完成对其他服务器的配置

```shell
cd oreo
cd .xxx/cmd/scripts/setup
./setup.sh
```

8. 在执行 Workload 之前, 需要手动配置各个服务器目前运行的数据库

- Postgres

```shell
docker run -d -p 5432:5432 --rm --name="apiary-postgres" --env POSTGRES_PASSWORD=dbos postgres:latest
```

1. 同步数据

```shell
rsync -avP root@49.233.36.106:~/oreo/benchmarks/cmd/data/ ~/Projects/oreo/benchmarks/cmd/data/


rsync -avP root@49.233.36.106:~/oreo ~/Projects/oreo2
```

## Run Evaluation

### YCSB

- Setup

```shell
./ycsb-setup.sh MongoDB1
./ycsb-setup.sh Redis

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
```

- Run

```shell
./realistic-full.sh -wl social -v -r
```

#### Order

- Setup

```shell
./realistic-setup.sh -wl order -id 2
```

- Run

```shell
./realistic-full.sh -wl order -v -r
```

### Optimization

- Setup

```shell
./opt-setup.sh -id 2
```

- Run

> 注意查看配置文件中的 TxnOperationGroup, zipfian 等参数

```shell
./opt-full.sh -wl RMW -v -r

./opt-full.sh -wl RRMW -v -r
```

### Read Strategy

- Setup

```shell
./read-setup.sh -id 2
```

- Run

```shell
./read-full.sh -wl RMW -v -r

./read-full.sh -wl RRMW -v -r
```

修改 `read-full.sh`，只运行 `readStrategy=p` 的部分，手动修改 `Chain Length`

- cache 数据

```shell

```
