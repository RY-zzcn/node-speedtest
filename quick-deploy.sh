#!/bin/bash

# 节点管理测速系统Docker快速部署脚本
# 版本: 1.0.0

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # 无颜色

# 默认配置
PANEL_PORT=8080
NODE_PORT=8081
INSTALL_DIR="$(pwd)/node-speedtest"
PANEL_URL="http://localhost:${PANEL_PORT}"
ADMIN_USERNAME="admin"
ADMIN_PASSWORD="admin"
SECRET_KEY=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)
NODE_NAME="本地节点"

# 检查Docker是否安装
check_docker() {
    echo -e "${BLUE}[信息] 检查Docker环境...${NC}"
    
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}[错误] Docker未安装，请先安装Docker${NC}"
        echo -e "${YELLOW}[提示] 可以使用以下命令安装Docker:${NC}"
        echo -e "curl -fsSL https://get.docker.com | sh"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}[错误] Docker Compose未安装，请先安装Docker Compose${NC}"
        echo -e "${YELLOW}[提示] 可以使用以下命令安装Docker Compose:${NC}"
        echo -e "curl -L \"https://github.com/docker/compose/releases/download/v2.15.1/docker-compose-\$(uname -s)-\$(uname -m)\" -o /usr/local/bin/docker-compose && chmod +x /usr/local/bin/docker-compose"
        exit 1
    fi
    
    echo -e "${GREEN}[成功] Docker环境检查通过${NC}"
}

# 创建目录结构
create_directories() {
    echo -e "${BLUE}[信息] 创建目录结构...${NC}"
    
    mkdir -p ${INSTALL_DIR}/data/panel
    mkdir -p ${INSTALL_DIR}/data/node
    mkdir -p ${INSTALL_DIR}/config
    mkdir -p ${INSTALL_DIR}/logs
    
    echo -e "${GREEN}[成功] 目录创建完成${NC}"
}

# 创建配置文件
create_config_files() {
    echo -e "${BLUE}[信息] 创建配置文件...${NC}"
    
    # 面板配置文件
    cat > ${INSTALL_DIR}/config/panel.json << EOF
{
  "listen_port": "${PANEL_PORT}",
  "database_path": "/app/data/panel.db",
  "log_path": "/app/logs/panel.log",
  "secret_key": "${SECRET_KEY}",
  "admin_username": "${ADMIN_USERNAME}",
  "admin_password": "${ADMIN_PASSWORD}",
  "panel_url": "${PANEL_URL}",
  "node_timeout": 120,
  "node_check_interval": 60,
  "speedtest_timeout": 300,
  "max_concurrent_tests": 5,
  "github_repo": "https://github.com/RY-zzcn/node-speedtest",
  "github_version": "v1.0.0"
}
EOF
    
    # 节点配置文件
    cat > ${INSTALL_DIR}/config/node.json << EOF
{
  "panel_url": "${PANEL_URL}",
  "node_key": "${SECRET_KEY}",
  "node_name": "${NODE_NAME}",
  "listen_port": "${NODE_PORT}",
  "log_path": "/app/logs/node.log",
  "data_dir": "/app/data",
  "heartbeat_interval": 30,
  "speedtest_timeout": 120,
  "download_threads": 4,
  "upload_threads": 2,
  "ping_count": 10
}
EOF
    
    echo -e "${GREEN}[成功] 配置文件创建完成${NC}"
}

# 创建Docker Compose文件
create_docker_compose_file() {
    echo -e "${BLUE}[信息] 创建Docker Compose文件...${NC}"
    
    cat > ${INSTALL_DIR}/docker-compose.yml << EOF
version: '3'

services:
  # 面板端服务
  panel:
    image: ry-zzcn/node-speedtest-panel:latest
    container_name: node-speedtest-panel
    restart: unless-stopped
    ports:
      - "${PANEL_PORT}:8080"
    volumes:
      - ./data/panel:/app/data
      - ./config/panel.json:/app/config.json
      - ./logs:/app/logs
    environment:
      - TZ=Asia/Shanghai
    networks:
      - node-speedtest-network

  # 节点端服务（本地节点）
  node:
    image: ry-zzcn/node-speedtest-node:latest
    container_name: node-speedtest-node
    restart: unless-stopped
    ports:
      - "${NODE_PORT}:8081"
    volumes:
      - ./data/node:/app/data
      - ./config/node.json:/app/config.json
      - ./logs:/app/logs
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
    
    echo -e "${GREEN}[成功] Docker Compose文件创建完成${NC}"
}

# 启动Docker容器
start_containers() {
    echo -e "${BLUE}[信息] 启动Docker容器...${NC}"
    
    cd ${INSTALL_DIR}
    docker-compose up -d
    
    # 检查容器状态
    if [ "$(docker ps -q -f name=node-speedtest-panel)" ]; then
        echo -e "${GREEN}[成功] 面板容器启动成功${NC}"
    else
        echo -e "${RED}[错误] 面板容器启动失败${NC}"
        echo -e "${YELLOW}[提示] 查看日志: docker logs node-speedtest-panel${NC}"
    fi
    
    if [ "$(docker ps -q -f name=node-speedtest-node)" ]; then
        echo -e "${GREEN}[成功] 节点容器启动成功${NC}"
    else
        echo -e "${RED}[错误] 节点容器启动失败${NC}"
        echo -e "${YELLOW}[提示] 查看日志: docker logs node-speedtest-node${NC}"
    fi
}

# 显示部署信息
show_deployment_info() {
    echo -e "\n${CYAN}========================================${NC}"
    echo -e "${CYAN}    节点管理测速系统部署完成    ${NC}"
    echo -e "${CYAN}========================================${NC}"
    
    echo -e "${GREEN}[面板信息]${NC}"
    echo -e "  面板访问地址: http://$(hostname -I | awk '{print $1}'):${PANEL_PORT}"
    echo -e "  管理员用户名: ${ADMIN_USERNAME}"
    echo -e "  管理员密码: ${ADMIN_PASSWORD}"
    
    echo -e "\n${GREEN}[节点信息]${NC}"
    echo -e "  本地节点已自动配置并连接到面板"
    
    echo -e "\n${YELLOW}[提示]${NC}"
    echo -e "  1. 请确保服务器防火墙已开放${PANEL_PORT}端口"
    echo -e "  2. 如需修改配置，请编辑${INSTALL_DIR}/config/目录下的配置文件后重启容器"
    echo -e "  3. 如需查看日志，请使用以下命令:"
    echo -e "     - 面板日志: docker logs node-speedtest-panel"
    echo -e "     - 节点日志: docker logs node-speedtest-node"
    echo -e "  4. 如需重启服务，请使用以下命令:"
    echo -e "     - cd ${INSTALL_DIR} && docker-compose restart"
    echo -e "  5. 如需停止服务，请使用以下命令:"
    echo -e "     - cd ${INSTALL_DIR} && docker-compose down"
    
    echo -e "\n${GREEN}[远程节点部署]${NC}"
    echo -e "  要在其他服务器上部署节点，请在面板中添加节点，然后使用生成的命令在远程服务器上安装"
    
    echo -e "${CYAN}========================================${NC}"
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
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
            --admin-username)
                ADMIN_USERNAME=$2
                shift 2
                ;;
            --admin-password)
                ADMIN_PASSWORD=$2
                shift 2
                ;;
            --install-dir)
                INSTALL_DIR=$2
                shift 2
                ;;
            --node-name)
                NODE_NAME=$2
                shift 2
                ;;
            --help)
                echo -e "${CYAN}节点管理测速系统Docker快速部署脚本${NC}"
                echo -e "${CYAN}用法:${NC}"
                echo -e "  $0 [选项]"
                echo -e ""
                echo -e "${CYAN}选项:${NC}"
                echo -e "  --panel-port PORT     指定面板端口（默认: 8080）"
                echo -e "  --node-port PORT      指定节点端口（默认: 8081）"
                echo -e "  --panel-url URL       指定面板URL（默认: http://localhost:面板端口）"
                echo -e "  --admin-username USER 指定管理员用户名（默认: admin）"
                echo -e "  --admin-password PASS 指定管理员密码（默认: admin）"
                echo -e "  --install-dir DIR     指定安装目录（默认: 当前目录/node-speedtest）"
                echo -e "  --node-name NAME      指定本地节点名称（默认: 本地节点）"
                echo -e "  --help                显示帮助信息"
                echo -e ""
                echo -e "${CYAN}示例:${NC}"
                echo -e "  $0 --panel-port 8080 --admin-username admin --admin-password password123"
                echo -e "  $0 --panel-url https://speedtest.example.com --node-name \"本地测试节点\""
                echo -e ""
                exit 0
                ;;
            *)
                echo -e "${RED}[错误] 未知参数: $1${NC}"
                exit 1
                ;;
        esac
    done
    
    # 如果没有指定面板URL，则使用默认值
    if [ "$PANEL_URL" = "http://localhost:${PANEL_PORT}" ]; then
        # 尝试获取服务器IP
        SERVER_IP=$(hostname -I | awk '{print $1}')
        if [ ! -z "$SERVER_IP" ]; then
            PANEL_URL="http://${SERVER_IP}:${PANEL_PORT}"
        fi
    fi
}

# 主函数
main() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}    节点管理测速系统Docker快速部署    ${NC}"
    echo -e "${CYAN}========================================${NC}"
    
    # 解析命令行参数
    parse_args "$@"
    
    # 显示配置信息
    echo -e "${BLUE}部署配置:${NC}"
    echo -e "  面板端口: ${PANEL_PORT}"
    echo -e "  节点端口: ${NODE_PORT}"
    echo -e "  面板URL: ${PANEL_URL}"
    echo -e "  安装目录: ${INSTALL_DIR}"
    echo -e "  管理员用户名: ${ADMIN_USERNAME}"
    echo -e "  管理员密码: ${ADMIN_PASSWORD}"
    echo -e "  本地节点名称: ${NODE_NAME}"
    echo -e "${CYAN}========================================${NC}"
    
    # 确认部署
    read -p "确认以上配置并开始部署？(y/n): " confirm
    if [[ $confirm != [yY] && $confirm != [yY][eE][sS] ]]; then
        echo -e "${YELLOW}[信息] 部署已取消${NC}"
        exit 0
    fi
    
    # 检查Docker
    check_docker
    
    # 创建目录
    create_directories
    
    # 创建配置文件
    create_config_files
    
    # 创建Docker Compose文件
    create_docker_compose_file
    
    # 启动容器
    start_containers
    
    # 显示部署信息
    show_deployment_info
}

# 执行主函数
main "$@" 