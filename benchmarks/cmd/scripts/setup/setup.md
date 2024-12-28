
1. 使用公网 IP 登录 Node1

```shell
ssh-copy-id -i ~/.ssh/id_ed25519.pub root@119.45.180.159
scp id_ed25519 root@公网IP:~/.ssh
scp go1.23.3.linux-amd64.tar.gz root@119.45.180.159:~
scp .tmux.conf root@119.45.180.159:~
```

2. 运行以下指令安装并运行 Docker

```shell
sudo yum install -y docker-ce

sudo systemctl start docker


cd /etc/yum.repos.d/
wget https://download.opensuse.org/repositories/shells:fish:release:3/CentOS_8/shells:fish:release:3.repo
yum install fish
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
yum install -y git,ripgrep


git clone git@github.com:KKKZOZ/oreo.git --depth=1

```

6. 安装 Golang

```shell
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz

vim .bashrc

export PATH=$PATH:/usr/local/go/bin

source .bashrc

go env -w GOPROXY=https://goproxy.cn,direct
```

7. 执行 `setup.sh` 完成对其他服务器的配置

```shell
cd oreo
cd .xxx/cmd/scripts/setup
./setup.sh
```

8. 在执行 Workload 之前, 需要手动配置各个服务器目前运行的数据库

10. 同步数据

```shell
rsync -avP root@119.45.180.159:~/oreo/benchmarks/cmd/data/ ~/Projects/oreo/benchmarks/cmd/data/


rsync -avP root@119.45.180.159:~/oreo ~/Projects/oreo2
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
