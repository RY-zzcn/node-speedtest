#!/bin/bash

# 节点管理测速系统 - 节点安装脚本
# 用法: curl -L https://your-panel-server.com/install.sh | bash -s -- YOUR_NODE_KEY [NODE_NAME]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # 无颜色

# 检查参数
if [ $# -lt 1 ]; then
    echo -e "${RED}错误: 缺少节点密钥参数${NC}"
    echo "用法: curl -L https://your-panel-server.com/install.sh | bash -s -- YOUR_NODE_KEY [NODE_NAME]"
    exit 1
fi

NODE_KEY=$1
NODE_NAME=${2:-$(hostname)}
PANEL_URL=${PANEL_URL:-"https://your-panel-server.com"}
INSTALL_DIR="/opt/node-speedtest"
CONFIG_FILE="${INSTALL_DIR}/config.json"
SERVICE_NAME="node-speedtest"

echo -e "${BLUE}=== 节点管理测速系统 - 节点安装脚本 ===${NC}"
echo -e "${BLUE}面板地址: ${PANEL_URL}${NC}"
echo -e "${BLUE}节点密钥: ${NODE_KEY}${NC}"
echo -e "${BLUE}节点名称: ${NODE_NAME}${NC}"
echo -e "${BLUE}安装目录: ${INSTALL_DIR}${NC}"

# 检查是否为root用户
if [ "$(id -u)" != "0" ]; then
    echo -e "${RED}错误: 此脚本需要root权限运行${NC}"
    echo "请使用sudo或以root用户身份运行此脚本"
    exit 1
fi

# 创建安装目录
echo -e "${YELLOW}创建安装目录...${NC}"
mkdir -p ${INSTALL_DIR}
mkdir -p ${INSTALL_DIR}/data
mkdir -p ${INSTALL_DIR}/logs

# 下载节点程序
echo -e "${YELLOW}下载节点程序...${NC}"
# TODO: 替换为实际的下载URL
# curl -L "${PANEL_URL}/downloads/node-speedtest" -o "${INSTALL_DIR}/node-speedtest"
# chmod +x "${INSTALL_DIR}/node-speedtest"

# 模拟下载
echo '#!/bin/bash
echo "节点测速程序模拟 - 版本1.0.0"
echo "节点密钥: '$NODE_KEY'"
echo "节点名称: '$NODE_NAME'"
echo "面板地址: '$PANEL_URL'"
sleep infinity
' > "${INSTALL_DIR}/node-speedtest"
chmod +x "${INSTALL_DIR}/node-speedtest"

# 创建配置文件
echo -e "${YELLOW}创建配置文件...${NC}"
cat > ${CONFIG_FILE} << EOF
{
  "panel_url": "${PANEL_URL}",
  "node_key": "${NODE_KEY}",
  "node_name": "${NODE_NAME}",
  "listen_port": "8081",
  "log_path": "${INSTALL_DIR}/logs/node.log",
  "data_dir": "${INSTALL_DIR}/data",
  "heartbeat_interval": 30,
  "speedtest_timeout": 120,
  "download_threads": 4,
  "upload_threads": 2,
  "ping_count": 10
}
EOF

# 创建systemd服务
echo -e "${YELLOW}创建系统服务...${NC}"
cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=节点管理测速系统 - 节点服务
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/node-speedtest -conf=${CONFIG_FILE}
Restart=always
RestartSec=10
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
echo -e "${YELLOW}启动服务...${NC}"
systemctl daemon-reload
systemctl enable ${SERVICE_NAME}
systemctl start ${SERVICE_NAME}

# 检查服务状态
echo -e "${YELLOW}检查服务状态...${NC}"
if systemctl is-active --quiet ${SERVICE_NAME}; then
    echo -e "${GREEN}节点服务已成功启动!${NC}"
else
    echo -e "${RED}节点服务启动失败，请检查日志文件: ${INSTALL_DIR}/logs/node.log${NC}"
    exit 1
fi

echo -e "${GREEN}=== 安装完成 ===${NC}"
echo -e "${GREEN}节点已成功安装并连接到面板服务器${NC}"
echo -e "${BLUE}配置文件: ${CONFIG_FILE}${NC}"
echo -e "${BLUE}日志文件: ${INSTALL_DIR}/logs/node.log${NC}"
echo -e "${BLUE}控制命令: ${NC}"
echo -e "${BLUE}  启动: systemctl start ${SERVICE_NAME}${NC}"
echo -e "${BLUE}  停止: systemctl stop ${SERVICE_NAME}${NC}"
echo -e "${BLUE}  重启: systemctl restart ${SERVICE_NAME}${NC}"
echo -e "${BLUE}  状态: systemctl status ${SERVICE_NAME}${NC}"
echo -e "${BLUE}  日志: journalctl -u ${SERVICE_NAME} -f${NC}"
echo ""
echo -e "${YELLOW}请在面板中检查节点是否成功连接${NC}" 