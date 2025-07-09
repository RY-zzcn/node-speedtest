# 节点管理测速系统部署教程

## 目录

- [系统介绍](#系统介绍)
- [部署方式](#部署方式)
  - [一键部署](#一键部署)
  - [手动部署](#手动部署)
  - [Docker部署](#docker部署)
- [配置说明](#配置说明)
  - [面板配置](#面板配置)
  - [节点配置](#节点配置)
- [系统管理](#系统管理)
  - [服务管理](#服务管理)
  - [日志查看](#日志查看)
  - [系统更新](#系统更新)
- [常见问题](#常见问题)
- [性能优化](#性能优化)
- [安全建议](#安全建议)

## 系统介绍

节点管理测速系统是一个用于管理和测试网络节点性能的系统，包括面板端和节点端两个主要组件。系统支持多平台部署，提供直观的Web界面，方便用户实时监控和管理多个测速节点。

### 系统架构

```
+-------------+        +-------------+        +-------------+
|   面板端    | <----> |   节点端1   | <----> |   节点端2   |
+-------------+        +-------------+        +-------------+
      ^                      ^                     ^
      |                      |                     |
+-------------+        +-------------+        +-------------+
|  管理员界面  |        |  测速目标1  |        |  测速目标2  |
+-------------+        +-------------+        +-------------+
```

### 系统要求

#### 面板端
- 操作系统：Linux (Debian/Ubuntu/CentOS等)
- CPU：1核以上
- 内存：512MB以上
- 存储：100MB以上
- 网络：公网IP或内网穿透

#### 节点端
- 操作系统：Linux (Debian/Ubuntu/CentOS等)
- CPU：1核以上
- 内存：256MB以上
- 存储：50MB以上
- 网络：可连接到面板端

## 部署方式

### 一键部署

一键部署脚本提供交互式界面，引导您完成整个安装过程，是最简单的部署方式。

```bash
# 下载部署脚本
curl -fsSL https://raw.githubusercontent.com/RY-zzcn/node-speedtest/main/quick-deploy.sh -o quick-deploy.sh

# 添加执行权限
chmod +x quick-deploy.sh

# 运行部署脚本
./quick-deploy.sh
```

脚本会引导您完成以下步骤：
1. 选择部署方式（面板+节点、仅面板、仅节点、Docker部署）
2. 配置面板URL、端口等基本信息
3. 设置管理员账号和密码
4. 自动下载和安装所需组件
5. 配置系统服务和自启动

### 手动部署

手动部署允许您更精细地控制安装过程和配置参数。

```bash
# 下载安装脚本
curl -fsSL https://raw.githubusercontent.com/RY-zzcn/node-speedtest/main/install.sh -o install.sh

# 添加执行权限
chmod +x install.sh

# 安装面板和节点
./install.sh --all --panel-url https://your-panel-domain.com --admin-username admin --admin-password yourpassword
```

#### 安装参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--panel` | 只安装面板端 | - |
| `--node` | 只安装节点端 | - |
| `--all` | 同时安装面板端和节点端 | 是 |
| `--docker` | 使用Docker部署 | 否 |
| `--panel-port PORT` | 指定面板端口 | 8080 |
| `--node-port PORT` | 指定节点端口 | 8081 |
| `--panel-url URL` | 指定面板URL（必须指定） | - |
| `--node-name NAME` | 指定节点名称 | 主机名 |
| `--admin-username USER` | 指定管理员用户名 | admin |
| `--admin-password PASS` | 指定管理员密码 | admin |

### Docker部署

Docker部署是最便捷的容器化部署方式，特别适合已经熟悉Docker的用户。

```bash
# 下载Docker Compose配置文件
curl -fsSL https://raw.githubusercontent.com/RY-zzcn/node-speedtest/main/docker-compose.yml -o docker-compose.yml

# 根据需要修改配置文件
nano docker-compose.yml

# 启动服务
docker-compose up -d
```

#### 自定义Docker配置

您可以通过修改`docker-compose.yml`文件来自定义Docker部署：

```yaml
version: '3'

services:
  # 面板端服务
  panel:
    image: ghcr.io/ry-zzcn/node-speedtest/node-speedtest-panel:latest
    container_name: node-speedtest-panel
    restart: unless-stopped
    ports:
      - "8080:8080"  # 修改为您需要的端口
    volumes:
      - ./data/panel:/app/data  # 数据持久化
      - ./logs/panel:/app/logs  # 日志持久化
      - ./config/panel.json:/app/config.json  # 配置文件
    environment:
      - TZ=Asia/Shanghai  # 时区设置

  # 节点端服务
  node:
    image: ghcr.io/ry-zzcn/node-speedtest/node-speedtest-node:latest
    container_name: node-speedtest-node
    restart: unless-stopped
    ports:
      - "8081:8081"  # 修改为您需要的端口
    volumes:
      - ./data/node:/app/data  # 数据持久化
      - ./logs/node:/app/logs  # 日志持久化
      - ./config/node.json:/app/config.json  # 配置文件
    environment:
      - TZ=Asia/Shanghai  # 时区设置
    depends_on:
      - panel
```

## 配置说明

### 面板配置

面板配置文件位于`/opt/node-speedtest/panel/config.json`（标准安装）或`./config/panel.json`（Docker安装）。

```json
{
  "listen_port": "8080",
  "database_path": "./data/panel.db",
  "log_path": "./logs/panel.log",
  "secret_key": "your_secret_key",
  "admin_username": "admin",
  "admin_password": "password",
  "panel_url": "https://your-panel-domain.com",
  "node_timeout": 120,
  "node_check_interval": 60,
  "speedtest_timeout": 300,
  "max_concurrent_tests": 5,
  "github_repo": "https://github.com/RY-zzcn/node-speedtest",
  "github_version": "v1.0.0"
}
```

#### 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `listen_port` | 面板监听端口 | 8080 |
| `database_path` | 数据库文件路径 | ./data/panel.db |
| `log_path` | 日志文件路径 | ./logs/panel.log |
| `secret_key` | 加密密钥，用于JWT等 | 随机生成 |
| `admin_username` | 管理员用户名 | admin |
| `admin_password` | 管理员密码 | admin |
| `panel_url` | 面板URL，用于节点连接 | - |
| `node_timeout` | 节点超时时间（秒） | 120 |
| `node_check_interval` | 节点检查间隔（秒） | 60 |
| `speedtest_timeout` | 测速超时时间（秒） | 300 |
| `max_concurrent_tests` | 最大并发测试数 | 5 |

### 节点配置

节点配置文件位于`/opt/node-speedtest/node/config.json`（标准安装）或`./config/node.json`（Docker安装）。

```json
{
  "listen_port": "8081",
  "log_path": "./logs/node.log",
  "panel_url": "https://your-panel-domain.com",
  "node_id": "",
  "node_key": "",
  "heartbeat_interval": 30,
  "download_threads": 4,
  "upload_threads": 2,
  "ping_count": 10
}
```

#### 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `listen_port` | 节点监听端口 | 8081 |
| `log_path` | 日志文件路径 | ./logs/node.log |
| `panel_url` | 面板URL | - |
| `node_id` | 节点ID，首次连接后自动生成 | - |
| `node_key` | 节点密钥，首次连接后自动生成 | - |
| `heartbeat_interval` | 心跳间隔（秒） | 30 |
| `download_threads` | 下载测试线程数 | 4 |
| `upload_threads` | 上传测试线程数 | 2 |
| `ping_count` | Ping测试次数 | 10 |

## 系统管理

### 服务管理

#### 标准安装

面板服务管理：
```bash
# 启动面板服务
systemctl start node-speedtest-panel

# 停止面板服务
systemctl stop node-speedtest-panel

# 重启面板服务
systemctl restart node-speedtest-panel

# 查看面板服务状态
systemctl status node-speedtest-panel
```

节点服务管理：
```bash
# 启动节点服务
systemctl start node-speedtest-node

# 停止节点服务
systemctl stop node-speedtest-node

# 重启节点服务
systemctl restart node-speedtest-node

# 查看节点服务状态
systemctl status node-speedtest-node
```

#### Docker安装

```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose down

# 重启所有服务
docker-compose restart

# 查看服务日志
docker-compose logs -f

# 仅启动面板服务
docker-compose up -d panel

# 仅启动节点服务
docker-compose up -d node
```

### 日志查看

#### 标准安装

```bash
# 查看面板日志
tail -f /opt/node-speedtest/logs/panel/panel.log

# 查看节点日志
tail -f /opt/node-speedtest/logs/node/node.log
```

#### Docker安装

```bash
# 查看面板日志
docker-compose logs -f panel

# 查看节点日志
docker-compose logs -f node
```

### 系统更新

#### 使用安装脚本更新

```bash
# 下载最新安装脚本
curl -fsSL https://raw.githubusercontent.com/RY-zzcn/node-speedtest/main/install.sh -o install.sh
chmod +x install.sh

# 更新面板和节点
./install.sh --all --update
```

#### Docker环境更新

```bash
# 拉取最新镜像
docker-compose pull

# 重启服务
docker-compose up -d
```

## 常见问题

### 1. 面板无法访问

**问题**：安装完成后无法访问面板Web界面。

**解决方案**：
- 检查服务器防火墙是否开放了面板端口（默认8080）
- 检查面板服务是否正常运行：`systemctl status node-speedtest-panel`
- 查看面板日志：`tail -f /opt/node-speedtest/logs/panel/panel.log`
- 确认配置文件中的`listen_port`设置正确

### 2. 节点无法连接到面板

**问题**：节点显示离线或无法连接到面板。

**解决方案**：
- 确认节点配置中的`panel_url`设置正确，必须包含协议（http://或https://）
- 检查节点和面板之间的网络连接：`ping your-panel-domain.com`
- 确认面板URL可以从节点服务器访问
- 查看节点日志：`tail -f /opt/node-speedtest/logs/node/node.log`

### 3. 测速结果异常

**问题**：测速结果不准确或超时失败。

**解决方案**：
- 检查节点服务器的网络状况
- 调整节点配置中的`download_threads`和`upload_threads`参数
- 增加`speedtest_timeout`参数值
- 确保测速目标服务器可以正常访问

### 4. 数据库错误

**问题**：面板报告数据库错误。

**解决方案**：
- 检查数据库文件权限：`chmod 644 /opt/node-speedtest/data/panel/panel.db`
- 确保数据目录可写：`chmod -R 755 /opt/node-speedtest/data`
- 如果数据库损坏，可以尝试备份后重新初始化：
  ```bash
  systemctl stop node-speedtest-panel
  mv /opt/node-speedtest/data/panel/panel.db /opt/node-speedtest/data/panel/panel.db.bak
  systemctl start node-speedtest-panel
  ```

## 性能优化

### 面板优化

1. **数据库优化**：
   - 定期清理过期数据：在Web界面中设置数据保留策略
   - 定期执行数据库压缩：`sqlite3 /opt/node-speedtest/data/panel/panel.db 'VACUUM;'`

2. **系统资源分配**：
   - 对于高负载场景，建议增加服务器内存和CPU资源
   - 调整`max_concurrent_tests`参数，避免过多并发测试

### 节点优化

1. **测速参数优化**：
   - 根据节点服务器性能调整`download_threads`和`upload_threads`
   - 高性能服务器可以适当增加线程数，低性能服务器应减少线程数

2. **网络优化**：
   - 确保节点服务器网络质量良好
   - 考虑使用专用网络接口进行测速

## 安全建议

1. **修改默认凭据**：
   - 立即修改默认管理员用户名和密码
   - 使用强密码，包含大小写字母、数字和特殊字符

2. **启用HTTPS**：
   - 为面板配置SSL证书，启用HTTPS访问
   - 可以使用Let's Encrypt免费证书

3. **访问控制**：
   - 限制面板管理接口的IP访问范围
   - 使用防火墙规则限制节点API的访问

4. **定期更新**：
   - 关注项目更新，及时升级到最新版本
   - 定期检查系统安全补丁

5. **数据备份**：
   - 定期备份数据库和配置文件
   - 设置自动备份计划 