# 节点管理测速系统一键部署说明

本文档详细介绍了节点管理测速系统的一键部署方法，包括面板部署和节点部署。

## 一键部署脚本

一键部署脚本(`install.sh`)是最简单的部署方式，它会自动完成以下操作：

1. 检测系统环境
2. 安装必要依赖
3. 创建目录结构
4. 下载最新版本的程序
5. 创建配置文件
6. 设置系统服务
7. 启动服务

### 部署面板和本地节点

```bash
# 下载一键部署脚本
curl -O https://raw.githubusercontent.com/RY-zzcn/node-speedtest/main/install.sh
chmod +x install.sh

# 部署面板和本地节点
./install.sh --all --panel-url https://your-panel-domain.com
```

### 只部署面板

```bash
./install.sh --panel --panel-url https://your-panel-domain.com --admin-username admin --admin-password your_password
```

### 只部署节点

```bash
./install.sh --node --panel-url https://your-panel-domain.com --node-name "香港节点"
```

## 参数说明

一键部署脚本支持以下参数：

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--panel` | 只安装面板 | - |
| `--node` | 只安装节点 | - |
| `--all` | 同时安装面板和节点 | 默认 |
| `--docker` | 使用Docker部署 | 否 |
| `--panel-port PORT` | 指定面板端口 | 8080 |
| `--node-port PORT` | 指定节点端口 | 8081 |
| `--panel-url URL` | 指定面板URL | 必须指定 |
| `--node-name NAME` | 指定节点名称 | 本机主机名 |
| `--admin-username USER` | 指定管理员用户名 | admin |
| `--admin-password PASS` | 指定管理员密码 | admin |
| `--help` | 显示帮助信息 | - |

## Docker快速部署

如果您熟悉Docker，可以使用Docker快速部署脚本(`quick-deploy.sh`)：

```bash
# 下载Docker快速部署脚本
curl -O https://raw.githubusercontent.com/RY-zzcn/node-speedtest/main/quick-deploy.sh
chmod +x quick-deploy.sh

# 使用默认配置部署
./quick-deploy.sh

# 或者自定义配置
./quick-deploy.sh --panel-port 8080 --admin-username admin --admin-password your_password --node-name "本地节点"
```

Docker快速部署脚本会自动完成以下操作：

1. 检查Docker环境
2. 创建目录结构
3. 创建配置文件
4. 创建Docker Compose文件
5. 启动容器

## 从面板部署节点

面板部署完成后，您可以通过面板界面添加节点，然后使用生成的命令在其他服务器上部署节点：

1. 登录面板（默认地址为 `http://面板IP:8080`，默认用户名和密码均为 `admin`）
2. 进入"节点管理"页面
3. 点击"添加节点"按钮，输入节点名称
4. 点击"生成节点"按钮，系统会生成节点密钥和安装命令
5. 复制生成的安装命令
6. 使用SSH连接到节点服务器，以root用户身份运行安装命令

```bash
# 从面板下载安装
curl -L https://your-panel-domain.com/api/install.sh | bash -s -- NODE_KEY NODE_NAME

# 从GitHub下载安装
curl -L https://your-panel-domain.com/api/install.sh | bash -s -- NODE_KEY NODE_NAME --github
```

## 部署后的操作

### 面板操作

面板部署完成后，您可以通过以下命令控制面板服务：

```bash
# 查看面板状态
systemctl status node-speedtest-panel

# 重启面板
systemctl restart node-speedtest-panel

# 停止面板
systemctl stop node-speedtest-panel

# 启动面板
systemctl start node-speedtest-panel

# 查看面板日志
journalctl -u node-speedtest-panel -f
```

如果使用Docker部署，则使用以下命令：

```bash
# 查看面板状态
docker ps | grep node-speedtest-panel

# 重启面板
docker restart node-speedtest-panel

# 停止面板
docker stop node-speedtest-panel

# 启动面板
docker start node-speedtest-panel

# 查看面板日志
docker logs -f node-speedtest-panel
```

### 节点操作

节点部署完成后，您可以通过以下命令控制节点服务：

```bash
# 查看节点状态
systemctl status node-speedtest

# 重启节点
systemctl restart node-speedtest

# 停止节点
systemctl stop node-speedtest

# 启动节点
systemctl start node-speedtest

# 查看节点日志
journalctl -u node-speedtest -f
```

如果使用Docker部署，则使用以下命令：

```bash
# 查看节点状态
docker ps | grep node-speedtest-node

# 重启节点
docker restart node-speedtest-node

# 停止节点
docker stop node-speedtest-node

# 启动节点
docker start node-speedtest-node

# 查看节点日志
docker logs -f node-speedtest-node
```

## 常见问题

### 1. 部署失败怎么办？

如果部署失败，请查看日志文件：

- 一键部署脚本日志：`/tmp/node-speedtest-install.log`
- 面板日志：`/opt/node-speedtest/logs/panel/panel.log`
- 节点日志：`/opt/node-speedtest/logs/node/node.log`

### 2. 如何修改配置？

- 面板配置文件：`/opt/node-speedtest/panel/config.json`
- 节点配置文件：`/opt/node-speedtest/node/config.json`

修改配置后需要重启相应服务：

```bash
systemctl restart node-speedtest-panel
systemctl restart node-speedtest
```

### 3. 如何更新系统？

使用一键部署脚本重新安装即可更新系统：

```bash
./install.sh --all --panel-url https://your-panel-domain.com
```

或者单独更新面板或节点：

```bash
./install.sh --panel --panel-url https://your-panel-domain.com
./install.sh --node --panel-url https://your-panel-domain.com --node-name "香港节点"
```

### 4. Docker部署的数据在哪里？

Docker部署的数据存储在以下位置：

- 面板数据：`./data/panel`
- 节点数据：`./data/node`
- 配置文件：`./config`
- 日志文件：`./logs`

### 5. 如何备份数据？

备份以下目录即可：

- 面板数据：`/opt/node-speedtest/data/panel`
- 节点数据：`/opt/node-speedtest/data/node`
- 配置文件：`/opt/node-speedtest/panel/config.json` 和 `/opt/node-speedtest/node/config.json`

如果使用Docker部署，备份整个安装目录即可。 