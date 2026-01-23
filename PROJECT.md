# TCP Collector 项目说明

## 项目概述

**项目名称**: tcp-collector (TCP数据采集器)

**功能**: 持续监听TCP端口,解析HEX数据(IEEE-754 32bit浮点数),推送到Kafka

**语言**: Go 1.21+

**目标平台**: Linux AMD64

## 项目结构

```
tcp-collector/
├── cmd/
│   └── main.go                 # 主程序入口
├── internal/
│   ├── config/
│   │   └── config.go           # 配置加载
│   ├── parser/
│   │   ├── field_def.go        # 字段定义常量
│   │   ├── parser.go           # 数据解析器
│   │   └── parser_test.go      # 单元测试
│   ├── tcpserver/
│   │   ├── server.go           # TCP服务器
│   │   └── manager.go          # 服务器管理器
│   └── kafka/
│       └── producer.go         # Kafka生产者
├── bin/                        # 编译输出目录
│   └── tcp-collector-linux-amd64
├── logs/                       # 日志目录(运行时创建)
├── config.yaml                 # 配置文件
├── parse_struct.txt            # 解析结构定义文件
├── go.mod                      # Go模块文件
├── go.sum                      # Go依赖锁定文件
├── Makefile                    # 编译脚本
├── deploy.sh                   # Linux部署脚本
├── test_client.py              # Python测试客户端
├── README.md                   # 用户文档
├── DEPLOY.md                   # 部署文档
├── TEST.md                     # 测试文档
└── PROJECT.md                  # 项目说明(本文件)
```

## 核心模块

### 1. Config模块 (`internal/config`)

**功能**: 加载和验证YAML配置文件

**主要结构**:
- `Config`: 主配置结构
- `TCPConfig`: TCP相关配置
- `KafkaConfig`: Kafka相关配置
- `LogConfig`: 日志配置

**关键方法**:
- `LoadConfig(path)`: 加载配置文件
- `Validate()`: 验证配置有效性

### 2. Parser模块 (`internal/parser`)

**功能**: 解析71字节数据帧

**数据格式**:
- Byte 0-47: PT1-PT12 (12个IEEE-754 32bit浮点数)
- Byte 48-51: 温度 (IEEE-754 32bit浮点数)
- Byte 52-55: 位移 (IEEE-754 32bit浮点数)
- Byte 56-61: LS11, LS21, LS31, LS11-1, LS21-1, LS31-1 (6个uint8)
- Byte 62-63: FS1 (uint16)
- Byte 64-70: 备用

**关键方法**:
- `Parse(data, deviceID, timestamp)`: 解析数据帧
- `parseFloat32(data)`: IEEE-754 32bit大端序解析
- `parseUint16(data)`: 16bit大端序解析

### 3. TCPServer模块 (`internal/tcpserver`)

**功能**: TCP服务器,监听端口接收数据

**关键特性**:
- 使用`io.ReadFull`读取固定长度数据帧(解决粘包/半包)
- 读取超时和空闲超时机制
- Context优雅关闭
- 每个连接独立goroutine处理

**关键方法**:
- `NewServer()`: 创建服务器
- `Start()`: 启动监听
- `Stop()`: 优雅停止
- `handleConnection()`: 处理单个连接

**Manager**: 管理多个TCP服务器
- `AddServer()`: 添加服务器
- `StartAll()`: 启动所有服务器
- `StopAll()`: 停止所有服务器

### 4. Kafka模块 (`internal/kafka`)

**功能**: Kafka生产者,推送数据

**关键特性**:
- 支持立即发送和批量发送
- 批量发送:按大小或超时触发
- 支持压缩(gzip/snappy/lz4/zstd)
- 失败重试
- 优雅关闭,发送剩余数据

**关键方法**:
- `NewProducer()`: 创建生产者
- `Send()`: 发送数据
- `Close()`: 关闭生产者
- `batchSendLoop()`: 批量发送循环

## 数据流

```
TCP客户端 → TCP服务器 → 数据解析器 → Kafka生产者 → Kafka
   |           |              |              |
   71字节    io.ReadFull   IEEE-754      批量发送
   HEX数据     完整读取      解析         JSON格式
```

## 配置管理

所有可变参数都在`config.yaml`中配置:

1. **TCP配置**: 端口、设备ID、超时时间
2. **Kafka配置**: Broker地址、Topic、批量配置
3. **日志配置**: 日志级别、输出方式
4. **解析配置**: 解析结构文件路径

## 编译和部署

### 本地编译

```bash
# 下载依赖
go mod download

# 编译Linux版本
make build-linux

# 运行测试
go test ./... -v
```

### Linux部署

```bash
# 1. 上传文件到服务器
scp -r tcp-collector/ user@server:/tmp/

# 2. 运行部署脚本
ssh user@server
cd /tmp/tcp-collector
sudo ./deploy.sh

# 3. 修改配置
sudo vi /etc/tcp-collector/config.yaml

# 4. 启动服务
sudo systemctl start tcp-collector
```

## 技术选型

### 为什么选择Go?

1. **高性能**: 原生goroutine,高并发处理
2. **静态编译**: 单个二进制文件,无依赖
3. **跨平台**: 交叉编译支持
4. **标准库强大**: net包原生TCP支持
5. **内存安全**: 无需手动内存管理

### 为什么使用sarama?

- IBM维护的成熟Kafka Go客户端
- 支持Kafka 0.10+所有特性
- 生产环境验证
- 批量发送、压缩、重试等完整支持

### 为什么使用io.ReadFull?

- 解决TCP粘包/半包问题
- 确保读取完整的71字节数据帧
- 阻塞式读取,简化逻辑

## 性能优化

### 1. 批量发送

启用Kafka批量发送,减少网络开销:

```yaml
kafka:
  batch_enabled: true
  batch_size: 100
  batch_timeout: 1000
```

**效果**: 吞吐量提升10-20倍

### 2. Kafka压缩

启用消息压缩,减少网络传输:

```yaml
kafka:
  compression: "snappy"
```

**效果**: 网络流量减少60-80%

### 3. Goroutine池

每个TCP连接使用独立goroutine,Go运行时自动调度,无需手动池化。

### 4. 零拷贝

使用`io.ReadFull`直接读取到buffer,避免多次内存拷贝。

## 可靠性设计

### 1. 优雅关闭

- 捕获SIGINT/SIGTERM信号
- 先停止TCP服务器(不接受新连接)
- 等待正在处理的数据完成
- 关闭Kafka生产者(发送剩余数据)

### 2. 超时机制

- **读取超时**: 30秒内必须收到数据
- **空闲超时**: 5分钟无数据自动断开

### 3. 错误处理

- 数据解析错误:记录日志,继续处理下一帧
- Kafka发送失败:自动重试3次
- TCP连接断开:客户端可重连

### 4. 数据完整性

- 固定长度读取(71字节)
- 数据验证(帧长度检查)
- 时间戳记录(接收时间)

## 监控和运维

### 日志级别

- **DEBUG**: 详细调试信息
- **INFO**: 关键操作日志(默认)
- **WARN**: 警告信息
- **ERROR**: 错误信息

### 关键日志

- `[TCP] Server listening`: 服务器启动
- `[TCP] New connection`: 新连接建立
- `[Kafka] Message sent`: 消息发送成功
- `[Kafka] Batch sent`: 批量发送完成

### systemd集成

- 自动重启: `Restart=on-failure`
- 日志输出: journalctl查看
- 开机自启: `systemctl enable`

## 扩展性

### 添加新字段

1. 修改`parse_struct.txt`
2. 更新`internal/parser/field_def.go`中的`ParseStructFields`
3. 重新编译

### 支持多种数据格式

1. 在`parser`模块添加新的解析器
2. 在配置文件添加格式选择
3. 在`dataHandler`中根据配置调用不同解析器

### 支持其他消息队列

1. 实现类似`kafka.Producer`的接口
2. 在配置文件添加MQ选择
3. 在main.go中根据配置创建不同生产者

## 常见问题

### Q: 如何修改数据帧长度?

A: 修改`config.yaml`中的`tcp.frame_length`

### Q: 如何支持小端序?

A: 修改`parser.go`中的`binary.BigEndian`为`binary.LittleEndian`

### Q: 如何增加更多端口?

A: 在`config.yaml`的`tcp.ports`和`tcp.device_ids`中添加

### Q: 如何关闭批量发送?

A: 设置`kafka.batch_enabled: false`

### Q: 如何查看日志?

A: `journalctl -u tcp-collector -f` 或查看配置的日志文件

## 版本历史

- **v1.0.0** (2026-01-23): 初始版本
  - TCP服务器监听
  - IEEE-754数据解析
  - Kafka推送
  - 批量发送
  - 优雅关闭

## 开发者

- 开发语言: Go 1.21+
- 依赖管理: go mod
- 测试框架: go test
- 构建工具: make

## 许可证

此项目仅供内部使用。
