# TCP数据采集器

## 功能说明
持续监听TCP端口,解析HEX数据并推送到Kafka

## 编译

```bash
# Linux AMD64
make build-linux

# 或直接使用go build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/tcp-collector cmd/main.go
```

## 配置文件

配置文件: `config.yaml`

### 主要配置项:
- `tcp.ports`: 监听的TCP端口列表
- `tcp.device_ids`: 对应每个端口的设备ID
- `kafka.brokers`: Kafka broker地址
- `kafka.topic`: 数据推送的topic
- `log.output`: 日志输出方式 (`console`/`file`/`both`)
- `log.file_path`: 日志文件路径(当output为file或both时)

## 运行

```bash
./tcp-collector -config config.yaml
```

## 日志文件

日志输出位置由配置文件决定:
- `console`: 仅输出到控制台
- `file`: 仅输出到文件
- `both`: 同时输出到控制台和文件(推荐)

日志文件特性:
- 自动轮转(按大小和天数)
- 自动压缩历史日志
- 默认路径: `/var/log/tcp-collector/collector.log`

## 数据格式

推送到Kafka的JSON格式:
```json
{
  "timestamp": 1737654321000,
  "device_id": "VK7015NH-001",
  "data": [
    {"quota": "PT1", "value": 632.123},
    {"quota": "PT2", "value": 632.123}
  ]
}
```

## 解析规则

- 前56字节: 14个IEEE-754 32bit浮点数(大端序)
  - PT1-PT12 (48字节)
  - 温度 (4字节)
  - 位移 (4字节)
- byte56-61: 6个单字节整数值
  - LS11, LS21, LS31, LS11-1, LS21-1, LS31-1
- byte62-63: FS1组合值(16bit整数,大端序)
- byte64-70: 备用字段
