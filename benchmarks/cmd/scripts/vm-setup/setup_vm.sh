#!/bin/bash

# 检查参数
if [ $# -ne 1 ]; then
    echo "Usage: $0 <public_ip>"
    exit 1
fi

PUBLIC_IP=$1
# PASSWORD="Qwert12345??"

# 确保 .ssh 目录存在
mkdir -p ~/.ssh

# 1. 配置免密登录到公网 IP
echo "Configuring passwordless login to public IP..."
sshpass -p "${PASSWORD}" ssh-copy-id -o StrictHostKeyChecking=no -i ~/.ssh/id_ed25519.pub root@${PUBLIC_IP}
scp ~/.ssh/otc_dse root@${PUBLIC_IP}:~/.ssh/

# 2. 登录并安装必要软件
echo "Installing required software..."
ssh root@${PUBLIC_IP} <<EOF
export PASSWORD="${PASSWORD}"
yum install -y docker-ce git ripgrep fish tmux sshpass
systemctl start docker

# 3. 配置内网机器免密登录
ssh-keygen -t rsa -N "" -f ~/.ssh/id_rsa
for ip in "10.206.206.3" "10.206.206.4" "10.206.206.5" "10.206.206.6"; do
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
scp ~/go1.23.3.linux-amd64.tar.gz ~/.tmux.conf root@${PUBLIC_IP}:~/
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go env -w GOPROXY=https://goproxy.cn,direct
EOF

echo "Setup completed successfully!"
