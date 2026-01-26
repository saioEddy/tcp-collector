#!/bin/bash
# 部署脚本

set -e

APP_NAME="tcp-collector"
INSTALL_DIR="/opt/${APP_NAME}"
CONFIG_DIR="/etc/${APP_NAME}"
LOG_DIR="/var/log/${APP_NAME}"
SERVICE_FILE="/etc/systemd/system/${APP_NAME}.service"

echo "=== TCP Collector 部署脚本 ==="

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then 
    echo "请使用root用户或sudo运行此脚本"
    exit 1
fi

# 创建目录
echo "创建目录..."
mkdir -p ${INSTALL_DIR}
mkdir -p ${CONFIG_DIR}
mkdir -p ${LOG_DIR}

# 复制二进制文件
echo "复制二进制文件..."
if [ ! -f "bin/${APP_NAME}-linux-amd64" ]; then
    echo "错误: 找不到编译后的二进制文件 bin/${APP_NAME}-linux-amd64"
    echo "请先运行: make build-linux"
    exit 1
fi

cp bin/${APP_NAME}-linux-amd64 ${INSTALL_DIR}/${APP_NAME}
chmod +x ${INSTALL_DIR}/${APP_NAME}

# 复制配置文件
echo "复制配置文件..."
if [ ! -f "${CONFIG_DIR}/config.yaml" ]; then
    cp config.yaml ${CONFIG_DIR}/config.yaml
    echo "配置文件已复制到 ${CONFIG_DIR}/config.yaml"
    echo "请根据实际情况修改配置文件!"
else
    echo "配置文件已存在,跳过复制"
fi

# 复制解析结构文件
if [ ! -f "${CONFIG_DIR}/parse_struct.txt" ]; then
    cp parse_struct.txt ${CONFIG_DIR}/parse_struct.txt
fi

# 创建systemd服务文件
echo "创建systemd服务..."
cat > ${SERVICE_FILE} << EOF
[Unit]
Description=TCP Data Collector Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/${APP_NAME} -config ${CONFIG_DIR}/config.yaml
Restart=on-failure
RestartSec=10
StandardOutput=append:${LOG_DIR}/stdout.log
StandardError=append:${LOG_DIR}/stderr.log

[Install]
WantedBy=multi-user.target
EOF

# 重新加载systemd
echo "重新加载systemd配置..."
systemctl daemon-reload

echo ""
echo "=== 部署完成 ==="
echo "安装目录: ${INSTALL_DIR}"
echo "配置目录: ${CONFIG_DIR}"
echo "日志目录: ${LOG_DIR}"
echo ""
echo "请先编辑配置文件: ${CONFIG_DIR}/config.yaml"
echo ""
echo "然后使用以下命令管理服务:"
echo "  启动服务: systemctl start ${APP_NAME}"
echo "  停止服务: systemctl stop ${APP_NAME}"
echo "  重启服务: systemctl restart ${APP_NAME}"
echo "  查看状态: systemctl status ${APP_NAME}"
echo "  开机自启: systemctl enable ${APP_NAME}"
echo "  查看日志: journalctl -u ${APP_NAME} -f"
echo "  应用日志: tail -f ${LOG_DIR}/collector.log"
