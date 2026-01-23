# TCP Collector 测试指南

## 单元测试

运行解析器测试:

```bash
go test ./internal/parser/ -v
```

预期输出:

```
=== RUN   TestParseFloat32
    parser_test.go:30: Parse result: 679.325806
--- PASS: TestParseFloat32 (0.00s)
=== RUN   TestParseFrame
    parser_test.go:87: Parse result: &{Timestamp:1737654321000 DeviceID:TEST-001 Data:[...]}
    parser_test.go:88: PT1: {Quota:PT1 Value:679.3258056640625}
--- PASS: TestParseFrame (0.00s)
PASS
```

## 本地集成测试

### 1. 启动本地Kafka

如果没有Kafka环境,可以使用Docker快速启动:

```bash
# 启动Kafka
docker run -d --name kafka \
  -p 9092:9092 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  -e KAFKA_LISTENERS=PLAINTEXT://0.0.0.0:9092 \
  apache/kafka:latest
```

### 2. 修改配置文件

确保`config.yaml`中的Kafka配置正确:

```yaml
kafka:
  brokers:
    - "localhost:9092"
  topic: "sensor_data"
```

### 3. 启动tcp-collector

```bash
# 本地运行(Mac/开发环境)
go run cmd/main.go -config config.yaml

# 或使用编译后的二进制
./bin/tcp-collector-linux-amd64 -config config.yaml
```

### 4. 发送测试数据

使用提供的Python测试客户端:

```bash
# 安装Python3 (如果没有)
# macOS: brew install python3
# Ubuntu: sudo apt install python3

# 发送测试数据
chmod +x test_client.py
./test_client.py --host localhost --port 9001 --count 10 --interval 1
```

参数说明:
- `--host`: 服务器地址
- `--port`: TCP端口
- `--count`: 发送次数
- `--interval`: 发送间隔(秒)

### 5. 验证Kafka消息

使用Kafka消费者查看消息:

```bash
# 使用Docker Kafka
docker exec -it kafka kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic sensor_data \
  --from-beginning

# 或使用kafkacat
kafkacat -b localhost:9092 -t sensor_data -C
```

预期输出JSON格式:

```json
{
  "timestamp": 1737654321000,
  "device_id": "VK7015NH-001",
  "data": [
    {"quota": "PT1", "value": 679.321},
    {"quota": "PT2", "value": 480.0},
    {"quota": "温度", "value": 25.5},
    {"quota": "位移", "value": 1.234},
    {"quota": "LS11", "value": 1},
    {"quota": "FS1", "value": 12345}
  ]
}
```

## 使用nc命令测试

如果没有Python环境,可以使用`nc`命令发送HEX数据:

```bash
# 创建测试数据文件(71字节)
# PT1: 44 29 D4 DA (679.321)
echo -n -e '\x44\x29\xD4\xDA' > test_frame.bin
# 填充剩余67字节为0
dd if=/dev/zero bs=67 count=1 >> test_frame.bin 2>/dev/null

# 发送到tcp-collector
nc localhost 9001 < test_frame.bin
```

## 压力测试

使用多个客户端并发发送:

```bash
# 启动10个并发客户端
for i in {1..10}; do
  ./test_client.py --host localhost --port 9001 --count 100 --interval 0.1 &
done

# 等待完成
wait
```

观察tcp-collector的日志输出和系统资源使用情况。

## 数据验证

### 验证IEEE-754解析

使用在线工具验证HEX到浮点数的转换:
- https://www.h-schmidt.net/FloatConverter/IEEE754.html

示例:
- HEX: `44 29 D4 DA`
- Binary: `01000100 00101001 11010100 11011010`
- Float: `679.325806...`

### 验证大端序

确保字节序正确:
- 大端序(Big Endian): 高位字节在前
- 示例: `0x1234` -> `0x12 0x34`

## 常见测试场景

### 1. 连接稳定性测试

```bash
# 长时间运行测试
./test_client.py --count 10000 --interval 0.5
```

### 2. 断线重连测试

```bash
# 发送一段时间后停止tcp-collector,再启动,观察客户端是否能重连
```

### 3. 批量发送性能测试

修改`config.yaml`:

```yaml
kafka:
  batch_enabled: true
  batch_size: 100
  batch_timeout: 1000
```

然后运行压力测试,观察批量发送效果。

### 4. 多端口测试

配置多个端口:

```yaml
tcp:
  ports: [9001, 9002, 9003]
  device_ids: ["DEV-001", "DEV-002", "DEV-003"]
```

同时向多个端口发送数据:

```bash
./test_client.py --port 9001 --count 100 &
./test_client.py --port 9002 --count 100 &
./test_client.py --port 9003 --count 100 &
```

## 故障注入测试

### 1. Kafka不可用

停止Kafka,观察tcp-collector的重试行为:

```bash
docker stop kafka
```

查看tcp-collector日志,应该看到重试错误日志。

### 2. 错误数据格式

发送非71字节的数据:

```bash
echo "invalid data" | nc localhost 9001
```

应该看到日志:`Invalid frame length`

### 3. 超时测试

建立连接但不发送数据,观察空闲超时:

```bash
nc localhost 9001
# 等待5分钟以上(默认idle_timeout=300秒)
```

应该看到日志:`Connection idle timeout`

## 性能基准

在典型硬件上(4核CPU, 8GB内存):

- **单连接吞吐量**: ~10,000 帧/秒
- **多连接(10个)吞吐量**: ~80,000 帧/秒
- **内存使用**: ~50MB
- **CPU使用**: ~30%

实际性能取决于:
- 硬件配置
- 网络延迟
- Kafka性能
- 是否启用批量发送
