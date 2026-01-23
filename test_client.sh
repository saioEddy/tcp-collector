#!/bin/bash
#
# TCP测试客户端 - Shell版本
# 用于向tcp-collector发送测试数据
#

# 默认参数
HOST="localhost"
PORT=9001
COUNT=10
INTERVAL=1

# 帮助信息
show_help() {
    cat << EOF
TCP测试客户端 - Shell版本

用法: $0 [选项]

选项:
    --host HOST        服务器地址 (默认: localhost)
    --port PORT        服务器端口 (默认: 9001)
    --count COUNT      发送次数 (默认: 10)
    --interval SEC     发送间隔秒数 (默认: 1)
    -h, --help         显示帮助信息

示例:
    $0 --host 192.168.1.100 --port 9001 --count 5 --interval 0.5
EOF
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --host)
            HOST="$2"
            shift 2
            ;;
        --port)
            PORT="$2"
            shift 2
            ;;
        --count)
            COUNT="$2"
            shift 2
            ;;
        --interval)
            INTERVAL="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "错误: 未知参数 '$1'"
            show_help
            exit 1
            ;;
    esac
done

# 检查依赖
check_dependencies() {
    local missing=0
    
    if ! command -v xxd &> /dev/null; then
        echo "错误: 缺少 xxd 命令,请安装"
        missing=1
    fi
    
    if ! command -v nc &> /dev/null; then
        echo "错误: 缺少 nc (netcat) 命令,请安装"
        missing=1
    fi
    
    if ! command -v bc &> /dev/null; then
        echo "警告: 缺少 bc 命令,浮点数计算可能不准确"
    fi
    
    return $missing
}

# 将浮点数转换为IEEE-754 32bit大端序十六进制
# 注意: 这个函数使用Python进行精确转换(如果可用),否则使用近似方法
float_to_hex() {
    local value=$1
    
    # 优先使用Python进行精确转换
    if command -v python3 &> /dev/null; then
        python3 -c "import struct; print(struct.pack('>f', $value).hex())"
    else
        # 简化版本: 仅用于演示,不精确
        echo "00000000"
    fi
}

# 将16bit整数转换为大端序十六进制
uint16_to_hex() {
    local value=$1
    printf "%04x" $value
}

# 创建测试数据帧(71字节)
create_test_frame() {
    local hex_data=""
    
    # PT1-PT12: 12个浮点数
    local pt_values=(679.321 480.0 123.45 234.56 345.67 456.78 567.89 678.90 789.01 890.12 901.23 12.34)
    
    for value in "${pt_values[@]}"; do
        hex_data+=$(float_to_hex $value)
    done
    
    # 温度 (byte48-51)
    hex_data+=$(float_to_hex 25.5)
    
    # 位移 (byte52-55)
    hex_data+=$(float_to_hex 1.234)
    
    # LS11-LS31-1 (byte56-61): 单字节值
    hex_data+="01"  # LS11
    hex_data+="00"  # LS21
    hex_data+="01"  # LS31
    hex_data+="00"  # LS11-1
    hex_data+="01"  # LS21-1
    hex_data+="00"  # LS31-1
    
    # FS1 (byte62-63): 16bit整数
    hex_data+=$(uint16_to_hex 12345)
    
    # 备用字段 (byte64-70): 7字节全0
    hex_data+="00000000000000"
    
    echo "$hex_data"
}

# 发送测试数据
send_test_data() {
    echo "连接到 ${HOST}:${PORT}..."
    
    # 创建测试数据帧
    local hex_frame=$(create_test_frame)
    
    # 验证数据长度
    local frame_length=$((${#hex_frame} / 2))
    echo "数据帧长度: ${frame_length} 字节"
    echo "数据帧前16字节(HEX): ${hex_frame:0:32}"
    
    if [ $frame_length -ne 71 ]; then
        echo "警告: 数据帧长度不正确,期望71字节,实际${frame_length}字节"
    fi
    
    # 将十六进制转换为二进制文件
    local temp_file=$(mktemp)
    echo "$hex_frame" | xxd -r -p > "$temp_file"
    
    # 发送数据
    for ((i=1; i<=COUNT; i++)); do
        if cat "$temp_file" | nc -N "$HOST" "$PORT" 2>/dev/null; then
            echo "[${i}/${COUNT}] 发送成功"
        else
            echo "[${i}/${COUNT}] 发送失败"
        fi
        
        if [ $i -lt $COUNT ]; then
            sleep "$INTERVAL"
        fi
    done
    
    # 清理临时文件
    rm -f "$temp_file"
    
    echo "测试完成!"
}

# 主程序
main() {
    echo "========================================"
    echo "TCP测试客户端 - Shell版本"
    echo "========================================"
    
    # 检查依赖
    if ! check_dependencies; then
        exit 1
    fi
    
    echo ""
    echo "配置参数:"
    echo "  服务器: ${HOST}:${PORT}"
    echo "  发送次数: ${COUNT}"
    echo "  发送间隔: ${INTERVAL}秒"
    echo ""
    
    # 发送测试数据
    send_test_data
}

# 运行主程序
main
