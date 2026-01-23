# TCP Collector 部署文档

## 系统要求

- 操作系统: Linux (CentOS 7+, Ubuntu 18.04+)
- 架构: x86_64 (AMD64)
- 依赖: 无(静态编译)
- Kafka: 0.10+ 版本

## 快速部署

### 1. 编译

在开发机上编译Linux二进制包:

```bash
make build-linux
```

编译完成后会生成: `bin/tcp-collector-linux-amd64`

### 2. 上传到服务器

将以下文件上传到服务器:

```bash
tcp-collector/
├── bin/tcp-collector-linux-amd64
├── config.yaml
├── parse_struct.txt
└── deploy.sh
```

### 3. 运行部署脚本

```bash
chmod +x deploy.sh
sudo ./deploy.sh
```

### 4. 修改配置文件

```bash
sudo vi /etc/tcp-collector/config.yaml
```

**必须修改的配置项:**

- `tcp.ports`: TCP监听端口列表
- `tcp.device_ids`: 对应每个端口的设备ID
- `kafka.brokers`: Kafka broker地址
- `kafka.topic`: Kafka topic名称

### 5. 启动服务

```bash
# 启动服务
sudo systemctl start tcp-collector

# 查看状态
sudo systemctl status tcp-collector

# 查看日志
sudo journalctl -u tcp-collector -f

# 设置开机自启
sudo systemctl enable tcp-collector
```

## 手动部署

如果不想使用systemd,可以手动运行:

```bash
./tcp-collector -config config.yaml
```

## 配置说明

### TCP配置

```yaml
tcp:
  ports: [9001, 9002]         # 监听端口列表
  device_ids:                  # 对应每个端口的设备ID
    - "VK7015NH-001"
    - "VK7015NH-002"
  frame_length: 71             # 数据帧长度(字节)
  read_timeout: 30             # 读取超时(秒)
  idle_timeout: 300            # 连接空闲超时(秒)
```

### Kafka配置

```yaml
kafka:
  brokers: ["localhost:9092"]  # Kafka broker地址
  topic: "sensor_data"         # Topic名称
  batch_enabled: true          # 是否启用批量发送
  batch_size: 100              # 批量大小
  batch_timeout: 1000          # 批量超时(毫秒)
  compression: "snappy"        # 压缩方式: none/gzip/snappy/lz4/zstd
  max_retries: 3               # 最大重试次数
```

### 批量发送说明

批量发送可以显著提高吞吐量,建议在高并发场景下启用:

- **batch_enabled**: true
- **batch_size**: 100 (达到100条消息时立即发送)
- **batch_timeout**: 1000 (1秒超时,即使不满100条也发送)

## 数据格式

推送到Kafka的JSON格式:

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

## 监控和日志

### 查看日志

```bash
# systemd日志
sudo journalctl -u tcp-collector -f

# 应用日志(如果配置了log.file_path)
tail -f /var/log/tcp-collector/collector.log
```

### 常见日志

- `[TCP] Server listening on 0.0.0.0:9001` - TCP服务器启动
- `[TCP] New connection from 192.168.1.100:12345` - 新连接
- `[TCP] Connection idle timeout` - 连接空闲超时
- `[Kafka] Message sent: partition=0, offset=12345` - 消息发送成功
- `[Kafka] Batch sent: 100 messages` - 批量发送成功

## 故障排查

### 1. 服务无法启动

```bash
# 查看详细日志
sudo journalctl -u tcp-collector -n 100

# 检查配置文件语法
cat /etc/tcp-collector/config.yaml
```

### 2. 无法连接Kafka

- 检查Kafka broker地址是否正确
- 检查网络连接: `telnet kafka-host 9092`
- 检查Kafka topic是否存在

### 3. TCP连接异常

- 检查端口是否被占用: `netstat -tlnp | grep 9001`
- 检查防火墙规则: `firewall-cmd --list-all`

### 4. 数据解析错误

- 检查数据帧长度是否为71字节
- 检查数据格式是否符合IEEE-754 32bit大端序

## 性能优化

### 1. 启用批量发送

```yaml
kafka:
  batch_enabled: true
  batch_size: 100
  batch_timeout: 1000
```

### 2. 启用Kafka压缩

```yaml
kafka:
  compression: "snappy"  # 推荐snappy,压缩率和性能平衡
```

### 3. 调整系统参数

```bash
# 增加文件描述符限制
ulimit -n 65535

# 增加TCP连接队列
sysctl -w net.core.somaxconn=4096
```

## 卸载

```bash
# 停止服务
sudo systemctl stop tcp-collector
sudo systemctl disable tcp-collector

# 删除服务文件
sudo rm /etc/systemd/system/tcp-collector.service
sudo systemctl daemon-reload

# 删除安装目录
sudo rm -rf /opt/tcp-collector
sudo rm -rf /etc/tcp-collector
sudo rm -rf /var/log/tcp-collector
```

## 技术支持

如有问题,请检查:

1. 配置文件是否正确
2. Kafka是否正常运行
3. 网络连接是否正常
4. 查看日志获取详细错误信息
