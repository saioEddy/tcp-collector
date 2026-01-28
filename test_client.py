#!/usr/bin/env python3
"""
TCP测试客户端
用于向tcp-collector发送测试数据
"""

import socket
import struct
import time

def float_to_bytes(value):
    """将浮点数转换为IEEE-754 32bit大端序字节"""
    return struct.pack('>f', value)

def uint16_to_bytes(value):
    """将16bit整数转换为大端序字节"""
    return struct.pack('>H', value)

def create_test_frame():
    """创建测试数据帧(70字节)"""
    # 原始代码(保留注释):
    # frame = bytearray(70)
    # 
    # # PT1-PT12: 12个浮点数
    # pt_values = [679.321, 480.0, 123.45, 234.56, 345.67, 456.78,
    #              567.89, 678.90, 789.01, 890.12, 901.23, 12.34]
    # 
    # for i, value in enumerate(pt_values):
    #     offset = i * 4
    #     frame[offset:offset+4] = float_to_bytes(value)
    # 
    # # 温度 (byte48-51)
    # frame[48:52] = float_to_bytes(25.5)
    # 
    # # 位移 (byte52-55)
    # frame[52:56] = float_to_bytes(1.234)
    # 
    # # LS11-LS31-1 (byte56-61): 单字节值
    # frame[56] = 1  # LS11
    # frame[57] = 0  # LS21
    # frame[58] = 1  # LS31
    # frame[59] = 0  # LS11-1
    # frame[60] = 1  # LS21-1
    # frame[61] = 0  # LS31-1
    # 
    # # FS1 (byte62-63): 16bit整数
    # frame[62:64] = uint16_to_bytes(12345)
    # 
    # # 备用字段(64-69)保持为0
    # 
    # return bytes(frame)
    
    # 新代码: 使用真实采集卡的数据格式 (70字节)
    # 数据来源: 网络调试助手截图
    hex_bytes = [
        0x41, 0xCA, 0xD4, 0x71, 0x41, 0x05, 0xA8, 0x71, 0x42, 0x47, 0x1A, 0xE4, 0x3F, 0x5F, 0x9C, 0x72,
        0x3F, 0x3F, 0xB8, 0xE4, 0x41, 0xCA, 0xD7, 0x8E, 0x3E, 0x5E, 0x71, 0xC7, 0x3D, 0x2E, 0x38, 0xE3,
        0x3E, 0x4A, 0x38, 0xE4, 0x41, 0xD5, 0x9F, 0x1D, 0x41, 0xD5, 0x5D, 0xC8, 0x41, 0x05, 0x38, 0x72,
        0x41, 0x4C, 0x1B, 0xD9, 0x44, 0x94, 0x24, 0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00  # 移除了最后1个字节,改为70字节
    ]
    
    frame = bytes(hex_bytes)
    
    if len(frame) != 70:
        print(f"警告: 数据长度为 {len(frame)} 字节,期望70字节")
    
    return frame

def send_test_data(host, port, count=10, interval=1):
    """发送测试数据"""
    print(f"连接到 {host}:{port}...")
    
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.connect((host, port))
        print("连接成功!")
        
        frame = create_test_frame()
        print(f"数据帧长度: {len(frame)} 字节")
        print(f"数据帧前16字节(HEX): {frame[:16].hex().upper()}")
        
        for i in range(count):
            sock.sendall(frame)
            print(f"[{i+1}/{count}] 发送成功")
            time.sleep(interval)
        
        print("测试完成!")
        sock.close()
        
    except Exception as e:
        print(f"错误: {e}")

if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="TCP测试客户端")
    parser.add_argument("--host", default="localhost", help="服务器地址")
    parser.add_argument("--port", type=int, default=9001, help="服务器端口")
    parser.add_argument("--count", type=int, default=10, help="发送次数")
    parser.add_argument("--interval", type=float, default=1.0, help="发送间隔(秒)")
    
    args = parser.parse_args()
    
    send_test_data(args.host, args.port, args.count, args.interval)
