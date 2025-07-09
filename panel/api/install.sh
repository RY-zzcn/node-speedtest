#!/bin/bash

# 节点管理测速系统节点安装脚本
# 版本: 1.0.0

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # 无颜色

# 默认值
PANEL_URL=""
NODE_KEY=""
NODE_NAME=""
USE_GITHUB=false
ARCH=$(uname -m)
INSTALL_DIR="/opt/node-speedtest"
NODE_DIR="${INSTALL_DIR}/node"
DATA_DIR="${INSTALL_DIR}/data/node"
LOGS_DIR="${INSTALL_DIR}/logs/node"
TEMP_DIR="/tmp/node-speedtest-install"
NODE_SERVICE="node-speedtest.service"
GITHUB_REPO="RY-zzcn/node-speedtest"
GITHUB_URL="https://github.com/${GITHUB_REPO}"
LOG_FILE="${TEMP_DIR}/install.log"

# 记录日志
log() {
    echo -e "$1" | tee -a ${LOG_FILE}
}

# 检测系统类型
check_system() {
    log "${BLUE}[信息] 检测系统环境...${NC}"
    
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
            log "${RED}[错误] 不支持的系统架构: $ARCH${NC}"
            exit 1
            ;;
    esac
    
    log "${GREEN}[成功] 系统类型: $OS, 架构: $ARCH${NC}"
}

# 安装依赖
install_dependencies() {
    log "${BLUE}[信息] 安装必要依赖...${NC}"
    
    case $OS in
        debian|ubuntu)
            apt update -y >> ${LOG_FILE} 2>&1
            apt install -y curl wget systemd >> ${LOG_FILE} 2>&1
            ;;
        centos|fedora|rhel)
            yum install -y curl wget systemd >> ${LOG_FILE} 2>&1
            ;;
        alpine)
            apk add --no-cache curl wget >> ${LOG_FILE} 2>&1
            ;;
        *)
            log "${YELLOW}[警告] 未知的操作系统类型，尝试安装基本依赖...${NC}"
            if command -v apt &> /dev/null; then
                apt update -y >> ${LOG_FILE} 2>&1
                apt install -y curl wget systemd >> ${LOG_FILE} 2>&1
            elif command -v yum &> /dev/null; then
                yum install -y curl wget systemd >> ${LOG_FILE} 2>&1
            elif command -v apk &> /dev/null; then
                apk add --no-cache curl wget >> ${LOG_FILE} 2>&1
            else
                log "${RED}[错误] 无法安装依赖，请手动安装curl和wget${NC}"
                exit 1
            fi
            ;;
    esac
    
    log "${GREEN}[成功] 依赖安装完成${NC}"
}

# 创建目录
create_directories() {
    log "${BLUE}[信息] 创建必要目录...${NC}"
    
    mkdir -p $NODE_DIR
    mkdir -p $DATA_DIR
    mkdir -p $LOGS_DIR
    mkdir -p $TEMP_DIR
    
    log "${GREEN}[成功] 目录创建完成${NC}"
}

# 下载节点程序
download_node() {
    log "${BLUE}[信息] 下载节点程序...${NC}"
    
    if [ "$USE_GITHUB" = true ]; then
        log "${BLUE}[信息] 从GitHub下载节点程序...${NC}"
        
        # 获取最新版本号
        LATEST_VERSION=$(curl -s https://api.github.com/repos/${GITHUB_REPO}/releases/latest | grep -o '"tag_name": ".*"' | sed 's/"tag_name": "//;s/"//g')
        
        if [ -z "$LATEST_VERSION" ]; then
            LATEST_VERSION="main"
            log "${YELLOW}[警告] 无法获取最新版本号，使用main分支${NC}"
        else
            log "${GREEN}[成功] 最新版本: $LATEST_VERSION${NC}"
        fi
        
        # 下载节点程序
        curl -L ${GITHUB_URL}/releases/download/${LATEST_VERSION}/node-${ARCH}.tar.gz -o ${TEMP_DIR}/node.tar.gz >> ${LOG_FILE} 2>&1
        
        if [ $? -ne 0 ]; then
            log "${RED}[错误] 从GitHub下载节点程序失败，尝试从面板下载...${NC}"
            curl -L ${PANEL_URL}/api/download/node-${ARCH}?key=${NODE_KEY} -o ${NODE_DIR}/node >> ${LOG_FILE} 2>&1
            
            if [ $? -ne 0 ]; then
                log "${RED}[错误] 下载节点程序失败，请检查网络连接和面板URL${NC}"
                exit 1
            fi
        else
            # 解压文件
            tar -xzf ${TEMP_DIR}/node.tar.gz -C ${NODE_DIR} >> ${LOG_FILE} 2>&1
        fi
    else
        log "${BLUE}[信息] 从面板下载节点程序...${NC}"
        curl -L ${PANEL_URL}/api/download/node-${ARCH}?key=${NODE_KEY} -o ${NODE_DIR}/node >> ${LOG_FILE} 2>&1
        
        if [ $? -ne 0 ]; then
            log "${RED}[错误] 从面板下载节点程序失败，请检查网络连接和面板URL${NC}"
            exit 1
        fi
    fi
    
    # 设置执行权限
    chmod +x ${NODE_DIR}/node
    
    log "${GREEN}[成功] 节点程序下载完成${NC}"
}

# 创建节点配置文件
create_node_config() {
    log "${BLUE}[信息] 创建节点配置文件...${NC}"
    
    # 创建配置文件
    cat > ${NODE_DIR}/config.json << EOF
{
  "panel_url": "${PANEL_URL}",
  "node_key": "${NODE_KEY}",
  "node_name": "${NODE_NAME}",
  "listen_port": "8081",
  "log_path": "${LOGS_DIR}/node.log",
  "data_dir": "${DATA_DIR}",
  "heartbeat_interval": 30,
  "speedtest_timeout": 120,
  "download_threads": 4,
  "upload_threads": 2,
  "ping_count": 10
}
EOF
    
    log "${GREEN}[成功] 节点配置文件创建完成${NC}"
}

# 创建systemd服务
create_systemd_service() {
    log "${BLUE}[信息] 创建systemd服务...${NC}"
    
    # 创建节点服务
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
    
    # 重新加载systemd
    systemctl daemon-reload >> ${LOG_FILE} 2>&1
    
    log "${GREEN}[成功] systemd服务创建完成${NC}"
}

# 启动节点服务
start_node_service() {
    log "${BLUE}[信息] 启动节点服务...${NC}"
    
    systemctl enable ${NODE_SERVICE} >> ${LOG_FILE} 2>&1
    systemctl start ${NODE_SERVICE} >> ${LOG_FILE} 2>&1
    
    # 检查服务状态
    if systemctl is-active --quiet ${NODE_SERVICE}; then
        log "${GREEN}[成功] 节点服务启动成功${NC}"
    else
        log "${RED}[错误] 节点服务启动失败，请检查日志${NC}"
        log "${YELLOW}[提示] 查看日志: journalctl -u ${NODE_SERVICE} -f${NC}"
    fi
}

# 清理临时文件
cleanup() {
    log "${BLUE}[信息] 清理临时文件...${NC}"
    rm -rf ${TEMP_DIR}
    log "${GREEN}[成功] 清理完成${NC}"
}

# 显示安装信息
show_install_info() {
    log "\n${CYAN}========================================${NC}"
    log "${CYAN}    节点管理测速系统节点安装完成    ${NC}"
    log "${CYAN}========================================${NC}"
    
    log "${GREEN}[节点信息]${NC}"
    log "  节点名称: ${NODE_NAME}"
    log "  节点密钥: ${NODE_KEY}"
    log "  面板地址: ${PANEL_URL}"
    log "  节点数据目录: ${DATA_DIR}"
    log "  节点日志目录: ${LOGS_DIR}"
    log "  节点配置文件: ${NODE_DIR}/config.json"
    log "  节点服务控制: systemctl [start|stop|restart|status] ${NODE_SERVICE}"
    
    log "\n${YELLOW}[提示]${NC}"
    log "  1. 请确保服务器防火墙已开放8081端口"
    log "  2. 如需修改配置，请编辑配置文件后重启服务"
    log "  3. 如需查看日志，请使用以下命令:"
    log "     - 节点日志: tail -f ${LOGS_DIR}/node.log"
    log "     - 或: journalctl -u ${NODE_SERVICE} -f"
    log "  4. 节点安装日志保存在: ${LOG_FILE}"
    
    log "${CYAN}========================================${NC}"
}

# 主函数
main() {
    # 创建临时目录和日志文件
    mkdir -p ${TEMP_DIR}
    touch ${LOG_FILE}
    
    # 显示欢迎信息
    log "${CYAN}========================================${NC}"
    log "${CYAN}    节点管理测速系统节点安装程序    ${NC}"
    log "${CYAN}========================================${NC}"
    
    # 解析命令行参数
    if [ $# -lt 2 ]; then
        log "${RED}[错误] 缺少必要参数${NC}"
        log "用法: $0 NODE_KEY NODE_NAME [--github]"
        exit 1
    fi
    
    NODE_KEY=$1
    NODE_NAME=$2
    
    # 检查是否使用GitHub下载
    if [ "$3" = "--github" ]; then
        USE_GITHUB=true
    fi
    
    # 获取面板URL
    PANEL_URL=$(echo "$0" | sed -E 's|^(https?://[^/]+)/.*|\1|')
    
    if [ -z "$PANEL_URL" ]; then
        log "${RED}[错误] 无法获取面板URL${NC}"
        exit 1
    fi
    
    log "${BLUE}安装信息:${NC}"
    log "  面板URL: ${PANEL_URL}"
    log "  节点密钥: ${NODE_KEY}"
    log "  节点名称: ${NODE_NAME}"
    log "  从GitHub下载: $([ "$USE_GITHUB" = true ] && echo "是" || echo "否")"
    log "${CYAN}========================================${NC}"
    
    # 检测系统
    check_system
    
    # 安装依赖
    install_dependencies
    
    # 创建目录
    create_directories
    
    # 下载节点程序
    download_node
    
    # 创建配置文件
    create_node_config
    
    # 创建服务
    create_systemd_service
    
    # 启动服务
    start_node_service
    
    # 清理临时文件
    cleanup
    
    # 显示安装信息
    show_install_info
}

# 执行主函数
main "$@" 2>&1 | tee -a ${LOG_FILE} 