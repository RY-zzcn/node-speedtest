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
    echo "用法: $0 [panel|node|all]"
    exit 1
fi

# 检查是否已编译
if [ ! -d "bin" ]; then
    echo -e "${YELLOW}未找到编译目录，正在编译...${NC}"
    make build
fi

# 根据参数启动相应的服务
case "$1" in
    panel)
        echo -e "${BLUE}启动面板端...${NC}"
        if [ -f "bin/panel" ]; then
            ./bin/panel
        else
            echo -e "${RED}面板端程序不存在，请先编译${NC}"
            exit 1
        fi
        ;;
    node)
        echo -e "${BLUE}启动节点端...${NC}"
        if [ -f "bin/node" ]; then
            ./bin/node
        else
            echo -e "${RED}节点端程序不存在，请先编译${NC}"
            exit 1
        fi
        ;;
    all)
        echo -e "${BLUE}同时启动面板端和节点端...${NC}"
        if [ -f "bin/panel" ] && [ -f "bin/node" ]; then
            # 后台启动面板端
            ./bin/panel &
            PANEL_PID=$!
            echo -e "${GREEN}面板端已启动，PID: ${PANEL_PID}${NC}"
            
            # 等待面板端启动完成
            sleep 2
            
            # 启动节点端
            ./bin/node
            
            # 当节点端退出时，也停止面板端
            kill $PANEL_PID
        else
            echo -e "${RED}程序不存在，请先编译${NC}"
            exit 1
        fi
        ;;
    *)
        echo -e "${RED}错误: 无效的参数 '$1'${NC}"
        echo "用法: $0 [panel|node|all]"
        exit 1
        ;;
esac

exit 0 