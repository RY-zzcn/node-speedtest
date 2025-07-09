# 节点管理测速系统

一个分布式的节点管理和网络测速系统，类似于哪吒探针，但专注于网络测速功能。

## 系统架构

```
+----------------+           +----------------+
|                |           |                |
|   面板服务器    | <-------> |   节点服务器1   |
|                |           |                |
+----------------+           +----------------+
        ^
        |                    +----------------+
        |                    |                |
        +-----------------> |   节点服务器2   |
                            |                |
                            +----------------+
                                    ...
                            +----------------+
                            |                |
                            |   节点服务器N   |
                            |                |
                            +----------------+
```

## 主要功能

- **节点管理**：集中管理多个节点，支持添加、删除、查看节点状态
- **网络测速**：测试节点之间的网络连接质量，包括延迟、下载速度和上传速度
- **数据可视化**：直观展示测速结果和节点状态
- **用户认证**：安全的用户登录和权限控制
- **API接口**：提供RESTful API接口，方便集成到其他系统
- **多种部署方式**：支持Docker部署和手动部署
- **灵活的节点安装**：支持从面板下载和从GitHub下载两种节点安装方式
- **一键部署**：提供简单的一键部署脚本，快速搭建整个系统

## 技术栈

- **后端**：Go语言
- **前端**：HTML/CSS/JavaScript
- **数据库**：SQLite
- **部署**：Docker, systemd

## 项目结构

```
node-speedtest/
├── docs/                    # 文档目录
│   ├── deployment.md        # 详细部署文档
│   └── one-click-deploy.md  # 一键部署文档
├── panel/                   # 面板服务端代码
│   ├── api/                 # API接口
│   │   ├── install.sh       # 节点安装脚本
│   │   └── handler.go       # API处理器
│   ├── config/              # 配置相关
│   ├── models/              # 数据模型
│   └── web/                 # Web界面
├── node/                    # 节点客户端代码
│   ├── api/                 # 节点API
│   ├── config/              # 节点配置
│   └── speedtest/           # 测速功能
├── install.sh               # 主安装脚本
├── quick-deploy.sh          # Docker快速部署脚本
├── docker-compose-deploy.yml # Docker部署配置
├── docker-compose.yml       # Docker开发配置
├── Dockerfile               # Docker构建文件
├── 部署教程.html             # HTML格式部署教程
└── README.md                # 项目说明
```

## 快速开始

### 一键部署（推荐）

使用一键部署脚本是最简单的方式，它会自动配置和启动所有必要的服务。

```bash
# 下载一键部署脚本
curl -O https://raw.githubusercontent.com/RY-zzcn/node-speedtest/main/install.sh
chmod +x install.sh

# 部署面板和本地节点
./install.sh --all

# 或者只部署面板
./install.sh --panel --panel-url https://your-panel-domain.com --admin-username admin --admin-password your_password

# 或者只部署节点
./install.sh --node --panel-url https://your-panel-domain.com --node-name "香港节点"
```

### Docker快速部署

如果您熟悉Docker，可以使用Docker快速部署脚本：

```bash
# 下载Docker快速部署脚本
curl -O https://raw.githubusercontent.com/RY-zzcn/node-speedtest/main/quick-deploy.sh
chmod +x quick-deploy.sh

# 使用默认配置部署
./quick-deploy.sh

# 或者自定义配置
./quick-deploy.sh --panel-port 8080 --admin-username admin --admin-password your_password --node-name "本地节点"
```

### 手动部署

#### Docker部署面板

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

### 节点部署

节点部署有两种方式：从面板下载安装和从GitHub下载安装。

#### 从面板下载安装

```bash
curl -L https://your-panel-domain.com/api/install.sh | bash -s -- NODE_KEY NODE_NAME
```

#### 从GitHub下载安装

```bash
curl -L https://your-panel-domain.com/api/install.sh | bash -s -- NODE_KEY NODE_NAME --github
```

## 详细文档

更多详细信息，请参阅[部署文档](docs/deployment.md)或查看[部署教程](部署教程.html)。

## 贡献指南

1. Fork 本仓库
2. 创建您的特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交您的更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 打开一个 Pull Request

## 许可证

MIT License 