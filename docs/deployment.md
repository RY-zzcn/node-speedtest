# 节点管理测速系统部署文档

## 目录

1. [面板部署](#面板部署)
   - [Docker部署](#docker部署)
   - [手动部署](#手动部署)
2. [节点部署](#节点部署)
   - [从面板下载安装](#从面板下载安装)
   - [从GitHub下载安装](#从github下载安装)
   - [手动部署](#节点手动部署)
3. [Nginx反向代理配置](#nginx反向代理配置)
4. [常见问题](#常见问题)

## 面板部署

### Docker部署

使用Docker是最简单的部署方式，只需要几个命令即可完成部署。

```bash
# 拉取镜像
docker pull ry-zzcn/node-speedtest-panel:latest

# 创建数据目录
mkdir -p /opt/node-speedtest-panel/data
mkdir -p /opt/node-speedtest-panel/logs

# 创建配置文件
cat > /opt/node-speedtest-panel/config.json << EOF
{
  "listen_port": "8080",
  "database_path": "/app/data/panel.db",
  "log_path": "/app/logs/panel.log",
  "secret_key": "change_this_to_a_random_string",
  "admin_username": "admin",
  "admin_password": "admin",
  "panel_url": "https://your-panel-domain.com",
  "node_timeout": 120,
  "node_check_interval": 60,
  "speedtest_timeout": 300,
  "max_concurrent_tests": 5,
  "github_repo": "https://github.com/RY-zzcn/node-speedtest",
  "github_version": "v1.0.0"
}
EOF

# 启动容器
docker run -d \
  --name node-speedtest-panel \
  -p 8080:8080 \
  -v /opt/node-speedtest-panel/config.json:/app/config.json \
  -v /opt/node-speedtest-panel/data:/app/data \
  -v /opt/node-speedtest-panel/logs:/app/logs \
  --restart always \
  ry-zzcn/node-speedtest-panel:latest
```

### 手动部署

如果您不想使用Docker，也可以手动部署面板。

#### 前提条件

- Go 1.18或更高版本
- Git

#### 部署步骤

```bash
# 克隆仓库
git clone https://github.com/RY-zzcn/node-speedtest.git
cd node-speedtest/panel

# 编译
go build -o panel

# 创建数据目录
mkdir -p data logs

# 创建配置文件
cat > config.json << EOF
{
  "listen_port": "8080",
  "database_path": "./data/panel.db",
  "log_path": "./logs/panel.log",
  "secret_key": "change_this_to_a_random_string",
  "admin_username": "admin",
  "admin_password": "admin",
  "panel_url": "https://your-panel-domain.com",
  "node_timeout": 120,
  "node_check_interval": 60,
  "speedtest_timeout": 300,
  "max_concurrent_tests": 5,
  "github_repo": "https://github.com/RY-zzcn/node-speedtest",
  "github_version": "v1.0.0"
}
EOF

# 创建systemd服务
sudo cat > /etc/systemd/system/node-speedtest-panel.service << EOF
[Unit]
Description=节点管理测速系统 - 面板服务
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$(pwd)
ExecStart=$(pwd)/panel
Restart=always
RestartSec=10
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable node-speedtest-panel
sudo systemctl start node-speedtest-panel

# 检查服务状态
sudo systemctl status node-speedtest-panel
```

## 节点部署

节点部署有两种方式：从面板下载安装和从GitHub下载安装。

### 从面板下载安装

这是推荐的安装方式，面板会自动生成安装命令。

1. 登录面板
2. 进入"节点管理"页面
3. 点击"添加节点"，输入节点名称
4. 复制生成的安装命令（从面板下载）
5. 在节点服务器上运行安装命令

```bash
curl -L https://your-panel-domain.com/api/install.sh | bash -s -- NODE_KEY NODE_NAME
```

### 从GitHub下载安装

如果节点服务器无法直接访问面板服务器，可以使用从GitHub下载的方式安装。

1. 登录面板
2. 进入"节点管理"页面
3. 点击"添加节点"，输入节点名称
4. 复制生成的安装命令（从GitHub下载）
5. 在节点服务器上运行安装命令

```bash
curl -L https://your-panel-domain.com/api/install.sh | bash -s -- NODE_KEY NODE_NAME --github
```

### 节点手动部署

如果您不想使用安装脚本，也可以手动部署节点。

#### 前提条件

- Go 1.18或更高版本
- Git

#### 部署步骤

```bash
# 克隆仓库
git clone https://github.com/RY-zzcn/node-speedtest.git
cd node-speedtest/node

# 编译
go build -o node-speedtest

# 创建数据目录
mkdir -p /opt/node-speedtest/data
mkdir -p /opt/node-speedtest/logs

# 移动可执行文件
mv node-speedtest /opt/node-speedtest/

# 创建配置文件
cat > /opt/node-speedtest/config.json << EOF
{
  "panel_url": "https://your-panel-domain.com",
  "node_key": "YOUR_NODE_KEY",
  "node_name": "YOUR_NODE_NAME",
  "listen_port": "8081",
  "log_path": "/opt/node-speedtest/logs/node.log",
  "data_dir": "/opt/node-speedtest/data",
  "heartbeat_interval": 30,
  "speedtest_timeout": 120,
  "download_threads": 4,
  "upload_threads": 2,
  "ping_count": 10
}
EOF

# 创建systemd服务
cat > /etc/systemd/system/node-speedtest.service << EOF
[Unit]
Description=节点管理测速系统 - 节点服务
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/node-speedtest
ExecStart=/opt/node-speedtest/node-speedtest -conf=/opt/node-speedtest/config.json
Restart=always
RestartSec=10
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
systemctl daemon-reload
systemctl enable node-speedtest
systemctl start node-speedtest

# 检查服务状态
systemctl status node-speedtest
```

## Nginx反向代理配置

如果您想使用域名访问面板，可以使用Nginx作为反向代理。

```nginx
server {
    listen 80;
    server_name your-panel-domain.com;
    
    # 重定向HTTP到HTTPS
    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl http2;
    server_name your-panel-domain.com;
    
    # SSL证书配置
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    # SSL参数
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    
    # 代理设置
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 常见问题

### 1. 节点无法连接到面板

- 检查节点配置文件中的`panel_url`是否正确
- 检查节点密钥是否正确
- 检查面板服务器防火墙是否开放了相应端口
- 检查节点服务器是否能够访问面板服务器

### 2. 面板无法启动

- 检查配置文件是否正确
- 检查日志文件中的错误信息
- 检查数据目录权限是否正确

### 3. 测速结果不准确

- 检查节点服务器网络状况
- 增加测速超时时间
- 调整下载/上传线程数

### 4. 安装脚本失败

- 检查服务器是否能够访问面板或GitHub
- 检查节点密钥是否正确
- 尝试使用另一种安装方式

### 5. 如何更新节点程序

- 使用安装脚本重新安装
- 或者手动下载新版本替换旧版本

### 6. 如何备份面板数据

- 备份`data`目录下的数据库文件
- 备份配置文件

### 7. 如何迁移面板

- 备份数据库和配置文件
- 在新服务器上安装面板
- 恢复数据库和配置文件 