#!/bin/bash

# 节点管理测速系统启动脚本

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # 无颜色

# 检查是否有参数
if [ $# -lt 1 ]; then
    echo -e "${RED}错误: 缺少参数${NC}"
    echo "用法: $0 [panel|node|all] [配置文件路径]"
    echo "例如: $0 panel /app/panel/config.json"
    exit 1
fi

# 服务类型
SERVICE_TYPE=$1

# 配置文件路径
CONFIG_PATH=""
if [ $# -ge 2 ]; then
    CONFIG_PATH=$2
fi

# 检查是否已编译
if [ ! -d "bin" ]; then
    echo -e "${YELLOW}未找到编译目录，正在编译...${NC}"
    make build
fi

# 准备配置文件参数
PANEL_CONFIG_PARAM=""
NODE_CONFIG_PARAM=""

if [ -n "$CONFIG_PATH" ]; then
    # 如果提供了配置文件路径
    if [ "$SERVICE_TYPE" = "panel" ] || [ "$SERVICE_TYPE" = "all" ]; then
        # 检查面板配置文件是否存在
        if [ ! -f "$CONFIG_PATH" ] && [ "$SERVICE_TYPE" = "panel" ]; then
            echo -e "${RED}错误: 面板配置文件不存在: $CONFIG_PATH${NC}"
            exit 1
        fi
        PANEL_CONFIG_PARAM="-config $CONFIG_PATH"
    elif [ "$SERVICE_TYPE" = "node" ]; then
        # 检查节点配置文件是否存在
        if [ ! -f "$CONFIG_PATH" ]; then
            echo -e "${RED}错误: 节点配置文件不存在: $CONFIG_PATH${NC}"
            exit 1
        fi
        NODE_CONFIG_PARAM="-config $CONFIG_PATH"
    fi
else
    # 使用默认配置文件路径
    if [ "$SERVICE_TYPE" = "panel" ] || [ "$SERVICE_TYPE" = "all" ]; then
        if [ -f "panel/config.json" ]; then
            PANEL_CONFIG_PARAM="-config panel/config.json"
        elif [ -f "/app/panel/config.json" ]; then
            PANEL_CONFIG_PARAM="-config /app/panel/config.json"
        fi
    fi
    
    if [ "$SERVICE_TYPE" = "node" ] || [ "$SERVICE_TYPE" = "all" ]; then
        if [ -f "node/config.json" ]; then
            NODE_CONFIG_PARAM="-config node/config.json"
        elif [ -f "/app/node/config.json" ]; then
            NODE_CONFIG_PARAM="-config /app/node/config.json"
        fi
    fi
fi

# 根据参数启动相应的服务
case "$SERVICE_TYPE" in
    panel)
        echo -e "${BLUE}启动面板端...${NC}"
        if [ -f "bin/panel" ]; then
            echo -e "${GREEN}使用配置参数: $PANEL_CONFIG_PARAM${NC}"
            ./bin/panel $PANEL_CONFIG_PARAM
        else
            echo -e "${RED}面板端程序不存在，请先编译${NC}"
            exit 1
        fi
        ;;
    node)
        echo -e "${BLUE}启动节点端...${NC}"
        if [ -f "bin/node" ]; then
            echo -e "${GREEN}使用配置参数: $NODE_CONFIG_PARAM${NC}"
            ./bin/node $NODE_CONFIG_PARAM
        else
            echo -e "${RED}节点端程序不存在，请先编译${NC}"
            exit 1
        fi
        ;;
    all)
        echo -e "${BLUE}同时启动面板端和节点端...${NC}"
        if [ -f "bin/panel" ] && [ -f "bin/node" ]; then
            # 后台启动面板端
            echo -e "${GREEN}使用面板配置参数: $PANEL_CONFIG_PARAM${NC}"
            ./bin/panel $PANEL_CONFIG_PARAM &
            PANEL_PID=$!
            echo -e "${GREEN}面板端已启动，PID: ${PANEL_PID}${NC}"
            
            # 等待面板端启动完成
            sleep 2
            
            # 启动节点端
            echo -e "${GREEN}使用节点配置参数: $NODE_CONFIG_PARAM${NC}"
            ./bin/node $NODE_CONFIG_PARAM
            
            # 当节点端退出时，也停止面板端
            kill $PANEL_PID
        else
            echo -e "${RED}程序不存在，请先编译${NC}"
            exit 1
        fi
        ;;
    *)
        echo -e "${RED}错误: 无效的参数 '$SERVICE_TYPE'${NC}"
        echo "用法: $0 [panel|node|all] [配置文件路径]"
        exit 1
        ;;
esac

exit 0 