#!/bin/bash

# 节点管理测速系统一键部署脚本
# 作者: AI Assistant
# 版本: 1.0.0

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # 无颜色

# 配置变量
GITHUB_REPO="RY-zzcn/node-speedtest"
GITHUB_URL="https://github.com/${GITHUB_REPO}"
GITHUB_API="https://api.github.com/repos/${GITHUB_REPO}"
GITHUB_RAW="https://raw.githubusercontent.com/${GITHUB_REPO}/main"
DEFAULT_PANEL_PORT=8080
DEFAULT_NODE_PORT=8081
INSTALL_DIR="/opt/node-speedtest"
PANEL_DIR="${INSTALL_DIR}/panel"
NODE_DIR="${INSTALL_DIR}/node"
DATA_DIR="${INSTALL_DIR}/data"
LOGS_DIR="${INSTALL_DIR}/logs"
TEMP_DIR="/tmp/node-speedtest-install"
CONFIG_FILE="${INSTALL_DIR}/config.json"
PANEL_SERVICE="node-speedtest-panel.service"
NODE_SERVICE="node-speedtest.service"

# 检测系统类型
check_system() {
    echo -e "${BLUE}[信息] 检测系统环境...${NC}"
    
    # 检测操作系统类型
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
    elif [ -f /etc/lsb-release ]; then
        . /etc/lsb-release
        OS=$DISTRIB_ID
    elif [ -f /etc/debian_version ]; then
        OS="debian"
    elif [ -f /etc/redhat-release ]; then
        OS="centos"
    else
        OS=$(uname -s)
    fi
    
    # 转换为小写
    OS=$(echo $OS | tr '[:upper:]' '[:lower:]')
    
    # 检测架构
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="arm"
            ;;
        *)
            echo -e "${RED}[错误] 不支持的系统架构: $ARCH${NC}"
            exit 1
            ;;
    esac
    
    echo -e "${GREEN}[成功] 系统类型: $OS, 架构: $ARCH${NC}"
}

# 安装依赖
install_dependencies() {
    echo -e "${BLUE}[信息] 安装必要依赖...${NC}"
    
    case $OS in
        debian|ubuntu)
            apt update -y
            apt install -y curl wget git unzip tar systemd
            ;;
        centos|fedora|rhel)
            yum install -y curl wget git unzip tar systemd
            ;;
        alpine)
            apk add --no-cache curl wget git unzip tar
            ;;
        *)
            echo -e "${YELLOW}[警告] 未知的操作系统类型，尝试安装基本依赖...${NC}"
            if command -v apt &> /dev/null; then
                apt update -y
                apt install -y curl wget git unzip tar systemd
            elif command -v yum &> /dev/null; then
                yum install -y curl wget git unzip tar systemd
            elif command -v apk &> /dev/null; then
                apk add --no-cache curl wget git unzip tar
            else
                echo -e "${RED}[错误] 无法安装依赖，请手动安装curl、wget、git、unzip、tar和systemd${NC}"
                exit 1
            fi
            ;;
    esac
    
    echo -e "${GREEN}[成功] 依赖安装完成${NC}"
}

# 创建目录
create_directories() {
    echo -e "${BLUE}[信息] 创建必要目录...${NC}"
    
    mkdir -p $PANEL_DIR
    mkdir -p $NODE_DIR
    mkdir -p $DATA_DIR/panel
    mkdir -p $DATA_DIR/node
    mkdir -p $LOGS_DIR/panel
    mkdir -p $LOGS_DIR/node
    mkdir -p $TEMP_DIR
    
    echo -e "${GREEN}[成功] 目录创建完成${NC}"
}

# 生成随机密钥
generate_secret_key() {
    cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1
}

# 下载最新版本
download_latest_release() {
    echo -e "${BLUE}[信息] 获取最新版本信息...${NC}"
    
    # 获取最新版本号
    LATEST_VERSION=$(curl -s ${GITHUB_API}/releases/latest | grep -o '"tag_name": ".*"' | sed 's/"tag_name": "//;s/"//g')
    
    if [ -z "$LATEST_VERSION" ]; then
        LATEST_VERSION="main"
        echo -e "${YELLOW}[警告] 无法获取最新版本号，使用main分支${NC}"
    else
        echo -e "${GREEN}[成功] 最新版本: $LATEST_VERSION${NC}"
    fi
    
    # 下载面板程序
    echo -e "${BLUE}[信息] 下载面板程序...${NC}"
    curl -L ${GITHUB_URL}/releases/download/${LATEST_VERSION}/panel-${ARCH}.tar.gz -o ${TEMP_DIR}/panel.tar.gz
    
    # 下载节点程序
    echo -e "${BLUE}[信息] 下载节点程序...${NC}"
    curl -L ${GITHUB_URL}/releases/download/${LATEST_VERSION}/node-${ARCH}.tar.gz -o ${TEMP_DIR}/node.tar.gz
    
    # 解压文件
    echo -e "${BLUE}[信息] 解压程序文件...${NC}"
    tar -xzf ${TEMP_DIR}/panel.tar.gz -C ${PANEL_DIR}
    tar -xzf ${TEMP_DIR}/node.tar.gz -C ${NODE_DIR}
    
    # 设置执行权限
    chmod +x ${PANEL_DIR}/panel
    chmod +x ${NODE_DIR}/node
    
    echo -e "${GREEN}[成功] 程序下载完成${NC}"
}

# 创建面板配置文件
create_panel_config() {
    echo -e "${BLUE}[信息] 创建面板配置文件...${NC}"
    
    # 生成随机密钥
    SECRET_KEY=$(generate_secret_key)
    
    # 创建配置文件
    cat > ${PANEL_DIR}/config.json << EOF
{
  "listen_port": "${PANEL_PORT}",
  "database_path": "${DATA_DIR}/panel/panel.db",
  "log_path": "${LOGS_DIR}/panel/panel.log",
  "secret_key": "${SECRET_KEY}",
  "admin_username": "${ADMIN_USERNAME}",
  "admin_password": "${ADMIN_PASSWORD}",
  "panel_url": "${PANEL_URL}",
  "node_timeout": 120,
  "node_check_interval": 60,
  "speedtest_timeout": 300,
  "max_concurrent_tests": 5,
  "github_repo": "${GITHUB_URL}",
  "github_version": "${LATEST_VERSION}"
}
EOF
    
    echo -e "${GREEN}[成功] 面板配置文件创建完成${NC}"
}

# 创建节点配置文件
create_node_config() {
    echo -e "${BLUE}[信息] 创建节点配置文件...${NC}"
    
    # 生成节点密钥
    NODE_KEY=$(generate_secret_key)
    
    # 创建配置文件
    cat > ${NODE_DIR}/config.json << EOF
{
  "panel_url": "${PANEL_URL}",
  "node_key": "${NODE_KEY}",
  "node_name": "${NODE_NAME}",
  "listen_port": "${NODE_PORT}",
  "log_path": "${LOGS_DIR}/node/node.log",
  "data_dir": "${DATA_DIR}/node",
  "heartbeat_interval": 30,
  "speedtest_timeout": 120,
  "download_threads": 4,
  "upload_threads": 2,
  "ping_count": 10
}
EOF
    
    echo -e "${GREEN}[成功] 节点配置文件创建完成${NC}"
    echo -e "${YELLOW}[提示] 节点密钥: ${NODE_KEY}${NC}"
    echo -e "${YELLOW}[提示] 请在面板中添加此节点，使用上述密钥${NC}"
}

# 创建systemd服务
create_systemd_service() {
    echo -e "${BLUE}[信息] 创建systemd服务...${NC}"
    
    # 创建面板服务
    if [ "$INSTALL_PANEL" = true ]; then
        cat > /etc/systemd/system/${PANEL_SERVICE} << EOF
[Unit]
Description=节点管理测速系统 - 面板服务
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${PANEL_DIR}
ExecStart=${PANEL_DIR}/panel
Restart=always
RestartSec=10
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
    fi
    
    # 创建节点服务
    if [ "$INSTALL_NODE" = true ]; then
        cat > /etc/systemd/system/${NODE_SERVICE} << EOF
[Unit]
Description=节点管理测速系统 - 节点服务
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${NODE_DIR}
ExecStart=${NODE_DIR}/node
Restart=always
RestartSec=10
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
    fi
    
    # 重新加载systemd
    systemctl daemon-reload
    
    echo -e "${GREEN}[成功] systemd服务创建完成${NC}"
}

# 启动服务
start_services() {
    echo -e "${BLUE}[信息] 启动服务...${NC}"
    
    # 启动面板服务
    if [ "$INSTALL_PANEL" = true ]; then
        systemctl enable ${PANEL_SERVICE}
        systemctl start ${PANEL_SERVICE}
        
        # 检查服务状态
        if systemctl is-active --quiet ${PANEL_SERVICE}; then
            echo -e "${GREEN}[成功] 面板服务启动成功${NC}"
        else
            echo -e "${RED}[错误] 面板服务启动失败，请检查日志${NC}"
            echo -e "${YELLOW}[提示] 查看日志: journalctl -u ${PANEL_SERVICE} -f${NC}"
        fi
    fi
    
    # 启动节点服务
    if [ "$INSTALL_NODE" = true ]; then
        systemctl enable ${NODE_SERVICE}
        systemctl start ${NODE_SERVICE}
        
        # 检查服务状态
        if systemctl is-active --quiet ${NODE_SERVICE}; then
            echo -e "${GREEN}[成功] 节点服务启动成功${NC}"
        else
            echo -e "${RED}[错误] 节点服务启动失败，请检查日志${NC}"
            echo -e "${YELLOW}[提示] 查看日志: journalctl -u ${NODE_SERVICE} -f${NC}"
        fi
    fi
}

# 清理临时文件
cleanup() {
    echo -e "${BLUE}[信息] 清理临时文件...${NC}"
    rm -rf ${TEMP_DIR}
    echo -e "${GREEN}[成功] 清理完成${NC}"
}

# 显示安装信息
show_install_info() {
    echo -e "\n${CYAN}========================================${NC}"
    echo -e "${CYAN}    节点管理测速系统安装完成    ${NC}"
    echo -e "${CYAN}========================================${NC}"
    
    if [ "$INSTALL_PANEL" = true ]; then
        echo -e "${GREEN}[面板信息]${NC}"
        echo -e "  面板访问地址: ${PANEL_URL}"
        echo -e "  管理员用户名: ${ADMIN_USERNAME}"
        echo -e "  管理员密码: ${ADMIN_PASSWORD}"
        echo -e "  面板数据目录: ${DATA_DIR}/panel"
        echo -e "  面板日志目录: ${LOGS_DIR}/panel"
        echo -e "  面板配置文件: ${PANEL_DIR}/config.json"
        echo -e "  面板服务控制: systemctl [start|stop|restart|status] ${PANEL_SERVICE}"
    fi
    
    if [ "$INSTALL_NODE" = true ]; then
        echo -e "\n${GREEN}[节点信息]${NC}"
        echo -e "  节点名称: ${NODE_NAME}"
        echo -e "  节点密钥: ${NODE_KEY}"
        echo -e "  节点数据目录: ${DATA_DIR}/node"
        echo -e "  节点日志目录: ${LOGS_DIR}/node"
        echo -e "  节点配置文件: ${NODE_DIR}/config.json"
        echo -e "  节点服务控制: systemctl [start|stop|restart|status] ${NODE_SERVICE}"
    fi
    
    echo -e "\n${YELLOW}[提示]${NC}"
    echo -e "  1. 请确保服务器防火墙已开放相应端口"
    echo -e "  2. 如需修改配置，请编辑对应的配置文件后重启服务"
    echo -e "  3. 如需查看日志，请使用以下命令:"
    
    if [ "$INSTALL_PANEL" = true ]; then
        echo -e "     - 面板日志: tail -f ${LOGS_DIR}/panel/panel.log"
        echo -e "     - 或: journalctl -u ${PANEL_SERVICE} -f"
    fi
    
    if [ "$INSTALL_NODE" = true ]; then
        echo -e "     - 节点日志: tail -f ${LOGS_DIR}/node/node.log"
        echo -e "     - 或: journalctl -u ${NODE_SERVICE} -f"
    fi
    
    echo -e "${CYAN}========================================${NC}"
}

# 安装Docker
install_docker() {
    echo -e "${BLUE}[信息] 安装Docker...${NC}"
    
    # 检查Docker是否已安装
    if command -v docker &> /dev/null; then
        echo -e "${GREEN}[成功] Docker已安装${NC}"
        return
    fi
    
    # 安装Docker
    case $OS in
        debian|ubuntu)
            apt update -y
            apt install -y apt-transport-https ca-certificates curl software-properties-common
            curl -fsSL https://download.docker.com/linux/$OS/gpg | apt-key add -
            add-apt-repository "deb [arch=$(dpkg --print-architecture)] https://download.docker.com/linux/$OS $(lsb_release -cs) stable"
            apt update -y
            apt install -y docker-ce docker-ce-cli containerd.io
            ;;
        centos|fedora|rhel)
            yum install -y yum-utils
            yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
            yum install -y docker-ce docker-ce-cli containerd.io
            ;;
        *)
            curl -fsSL https://get.docker.com | sh
            ;;
    esac
    
    # 启动Docker
    systemctl enable docker
    systemctl start docker
    
    # 安装Docker Compose
    curl -L "https://github.com/docker/compose/releases/download/v2.15.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    
    echo -e "${GREEN}[成功] Docker安装完成${NC}"
}

# 使用Docker部署
deploy_with_docker() {
    echo -e "${BLUE}[信息] 使用Docker部署...${NC}"
    
    # 创建docker-compose.yml文件
    cat > ${INSTALL_DIR}/docker-compose.yml << EOF
version: '3'

services:
  # 面板端服务
  panel:
    image: ${GITHUB_REPO}/node-speedtest-panel:latest
    container_name: node-speedtest-panel
    restart: unless-stopped
    ports:
      - "${PANEL_PORT}:8080"
    volumes:
      - ${DATA_DIR}/panel:/app/data
      - ${PANEL_DIR}/config.json:/app/config.json
    environment:
      - TZ=Asia/Shanghai
    networks:
      - node-speedtest-network

  # 节点端服务
  node:
    image: ${GITHUB_REPO}/node-speedtest-node:latest
    container_name: node-speedtest-node
    restart: unless-stopped
    ports:
      - "${NODE_PORT}:8081"
    volumes:
      - ${DATA_DIR}/node:/app/data
      - ${NODE_DIR}/config.json:/app/config.json
    environment:
      - TZ=Asia/Shanghai
    depends_on:
      - panel
    networks:
      - node-speedtest-network

networks:
  node-speedtest-network:
    driver: bridge
EOF
    
    # 启动容器
    cd ${INSTALL_DIR}
    docker-compose up -d
    
    # 检查容器状态
    if [ "$(docker ps -q -f name=node-speedtest-panel)" ]; then
        echo -e "${GREEN}[成功] 面板容器启动成功${NC}"
    else
        echo -e "${RED}[错误] 面板容器启动失败，请检查日志${NC}"
        echo -e "${YELLOW}[提示] 查看日志: docker logs node-speedtest-panel${NC}"
    fi
    
    if [ "$(docker ps -q -f name=node-speedtest-node)" ]; then
        echo -e "${GREEN}[成功] 节点容器启动成功${NC}"
    else
        echo -e "${RED}[错误] 节点容器启动失败，请检查日志${NC}"
        echo -e "${YELLOW}[提示] 查看日志: docker logs node-speedtest-node${NC}"
    fi
}

# 显示帮助信息
show_help() {
    echo -e "${CYAN}节点管理测速系统一键部署脚本${NC}"
    echo -e "${CYAN}用法:${NC}"
    echo -e "  $0 [选项]"
    echo -e ""
    echo -e "${CYAN}选项:${NC}"
    echo -e "  -h, --help            显示帮助信息"
    echo -e "  -p, --panel           仅安装面板"
    echo -e "  -n, --node            仅安装节点"
    echo -e "  -a, --all             同时安装面板和节点（默认）"
    echo -e "  -d, --docker          使用Docker部署"
    echo -e "  --panel-port PORT     指定面板端口（默认: 8080）"
    echo -e "  --node-port PORT      指定节点端口（默认: 8081）"
    echo -e "  --panel-url URL       指定面板URL（必须指定）"
    echo -e "  --node-name NAME      指定节点名称（默认: 本机主机名）"
    echo -e "  --admin-username USER 指定管理员用户名（默认: admin）"
    echo -e "  --admin-password PASS 指定管理员密码（默认: admin）"
    echo -e ""
    echo -e "${CYAN}示例:${NC}"
    echo -e "  $0 --panel --panel-url https://speedtest.example.com --admin-username admin --admin-password password123"
    echo -e "  $0 --node --panel-url https://speedtest.example.com --node-name \"香港节点\""
    echo -e "  $0 --all --docker --panel-url https://speedtest.example.com"
    echo -e ""
}

# 主函数
main() {
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -p|--panel)
                INSTALL_PANEL=true
                INSTALL_NODE=false
                shift
                ;;
            -n|--node)
                INSTALL_PANEL=false
                INSTALL_NODE=true
                shift
                ;;
            -a|--all)
                INSTALL_PANEL=true
                INSTALL_NODE=true
                shift
                ;;
            -d|--docker)
                USE_DOCKER=true
                shift
                ;;
            --panel-port)
                PANEL_PORT=$2
                shift 2
                ;;
            --node-port)
                NODE_PORT=$2
                shift 2
                ;;
            --panel-url)
                PANEL_URL=$2
                shift 2
                ;;
            --node-name)
                NODE_NAME=$2
                shift 2
                ;;
            --admin-username)
                ADMIN_USERNAME=$2
                shift 2
                ;;
            --admin-password)
                ADMIN_PASSWORD=$2
                shift 2
                ;;
            *)
                echo -e "${RED}[错误] 未知参数: $1${NC}"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 设置默认值
    INSTALL_PANEL=${INSTALL_PANEL:-true}
    INSTALL_NODE=${INSTALL_NODE:-true}
    USE_DOCKER=${USE_DOCKER:-false}
    PANEL_PORT=${PANEL_PORT:-$DEFAULT_PANEL_PORT}
    NODE_PORT=${NODE_PORT:-$DEFAULT_NODE_PORT}
    NODE_NAME=${NODE_NAME:-$(hostname)}
    ADMIN_USERNAME=${ADMIN_USERNAME:-"admin"}
    ADMIN_PASSWORD=${ADMIN_PASSWORD:-"admin"}
    
    # 检查必要参数
    if [ -z "$PANEL_URL" ]; then
        echo -e "${RED}[错误] 必须指定面板URL（--panel-url）${NC}"
        show_help
        exit 1
    fi
    
    # 显示安装信息
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}    节点管理测速系统安装程序    ${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo -e "${BLUE}安装选项:${NC}"
    echo -e "  安装面板: $([ "$INSTALL_PANEL" = true ] && echo "是" || echo "否")"
    echo -e "  安装节点: $([ "$INSTALL_NODE" = true ] && echo "是" || echo "否")"
    echo -e "  使用Docker: $([ "$USE_DOCKER" = true ] && echo "是" || echo "否")"
    echo -e "  面板端口: ${PANEL_PORT}"
    echo -e "  节点端口: ${NODE_PORT}"
    echo -e "  面板URL: ${PANEL_URL}"
    echo -e "  节点名称: ${NODE_NAME}"
    echo -e "  管理员用户名: ${ADMIN_USERNAME}"
    echo -e "  管理员密码: ${ADMIN_PASSWORD}"
    echo -e "${CYAN}========================================${NC}"
    
    # 确认安装
    read -p "确认安装配置？(y/n): " confirm
    if [[ $confirm != [yY] && $confirm != [yY][eE][sS] ]]; then
        echo -e "${YELLOW}[信息] 安装已取消${NC}"
        exit 0
    fi
    
    # 检查系统
    check_system
    
    # 安装依赖
    install_dependencies
    
    # 创建目录
    create_directories
    
    # 如果使用Docker
    if [ "$USE_DOCKER" = true ]; then
        # 安装Docker
        install_docker
        
        # 创建面板配置文件
        if [ "$INSTALL_PANEL" = true ]; then
            create_panel_config
        fi
        
        # 创建节点配置文件
        if [ "$INSTALL_NODE" = true ]; then
            create_node_config
        fi
        
        # 使用Docker部署
        deploy_with_docker
    else
        # 下载最新版本
        download_latest_release
        
        # 创建面板配置文件
        if [ "$INSTALL_PANEL" = true ]; then
            create_panel_config
        fi
        
        # 创建节点配置文件
        if [ "$INSTALL_NODE" = true ]; then
            create_node_config
        fi
        
        # 创建systemd服务
        create_systemd_service
        
        # 启动服务
        start_services
    fi
    
    # 清理临时文件
    cleanup
    
    # 显示安装信息
    show_install_info
}

# 执行主函数
main "$@" 