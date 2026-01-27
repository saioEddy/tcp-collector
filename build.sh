#!/bin/bash

# TCP Collector 编译打包脚本 (仅 Linux 版本)

set -e  # 遇到错误立即退出

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== TCP Collector 编译打包 (Linux) ===${NC}"

# 项目信息
PROJECT_NAME="tcp-collector"
MAIN_FILE="./cmd/main.go"
BIN_DIR="./bin"

# 清理并创建 bin 目录
echo -e "${YELLOW}清理并创建 bin 目录...${NC}"
rm -rf ${BIN_DIR}
mkdir -p ${BIN_DIR}

# 编译 Linux AMD64 版本
echo -e "${YELLOW}编译 Linux AMD64 版本...${NC}"
GOOS=linux GOARCH=amd64 go build -o ${BIN_DIR}/${PROJECT_NAME}-linux-amd64 ${MAIN_FILE}
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Linux AMD64 版本编译成功${NC}"
    ls -lh ${BIN_DIR}/${PROJECT_NAME}-linux-amd64
else
    echo -e "${RED}✗ Linux AMD64 版本编译失败${NC}"
    exit 1
fi

# 显示编译结果
echo -e "\n${GREEN}=== 编译完成 ===${NC}"
echo -e "${YELLOW}文件列表:${NC}"
ls -lh ${BIN_DIR}/

# 计算总大小
TOTAL_SIZE=$(du -sh ${BIN_DIR} | awk '{print $1}')
echo -e "\n${GREEN}总大小: ${TOTAL_SIZE}${NC}"

echo -e "\n${GREEN}=== 打包完成 ===${NC}"
echo -e "${YELLOW}提示: 上传到 Linux 服务器时使用:${NC}"
echo -e "  scp ${BIN_DIR}/${PROJECT_NAME}-linux-amd64 root@adamdemo:/opt/tcp_test/"
