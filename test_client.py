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
    """创建测试数据帧(71字节)"""
    frame = bytearray(71)
    
    # PT1-PT12: 12个浮点数
    pt_values = [679.321, 480.0, 123.45, 234.56, 345.67, 456.78,
                 567.89, 678.90, 789.01, 890.12, 901.23, 12.34]
    
    for i, value in enumerate(pt_values):
        offset = i * 4
        frame[offset:offset+4] = float_to_bytes(value)
    
    # 温度 (byte48-51)
    frame[48:52] = float_to_bytes(25.5)
    
    # 位移 (byte52-55)
    frame[52:56] = float_to_bytes(1.234)
    
    # LS11-LS31-1 (byte56-61): 单字节值
    frame[56] = 1  # LS11
    frame[57] = 0  # LS21
    frame[58] = 1  # LS31
    frame[59] = 0  # LS11-1
    frame[60] = 1  # LS21-1
    frame[61] = 0  # LS31-1
    
    # FS1 (byte62-63): 16bit整数
    frame[62:64] = uint16_to_bytes(12345)
    
    # 备用字段保持为0
    
    return bytes(frame)

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
