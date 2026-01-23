package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tcp-collector/internal/config"
	"tcp-collector/internal/kafka"
	"tcp-collector/internal/parser"
	"tcp-collector/internal/tcpserver"
)

var (
	configPath = flag.String("config", "config.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Load config error: %v", err)
	}

	log.Printf("Config loaded: %d TCP ports, Kafka brokers: %v", len(cfg.TCP.Ports), cfg.Kafka.Brokers)

	// 创建解析器
	dataParser := parser.NewParser()
	log.Printf("Parser created")

	// 创建Kafka生产者
	kafkaProducer, err := kafka.NewProducer(&cfg.Kafka)
	if err != nil {
		log.Fatalf("Create kafka producer error: %v", err)
	}
	defer kafkaProducer.Close()

	// 创建数据处理函数
	dataHandler := func(deviceID string, data []byte, timestamp int64) error {
		// 解析数据
		parsed, err := dataParser.Parse(data, deviceID, timestamp)
		if err != nil {
			log.Printf("[Handler] Parse data error: %v", err)
			return err
		}

		// 发送到Kafka
		if err := kafkaProducer.Send(parsed); err != nil {
			log.Printf("[Handler] Send to kafka error: %v", err)
			return err
		}

		return nil
	}

	// 创建TCP服务器管理器
	serverManager := tcpserver.NewManager()

	// 为每个端口创建TCP服务器
	for i, port := range cfg.TCP.Ports {
		deviceID := cfg.TCP.DeviceIDs[i]
		server := tcpserver.NewServer(
			port,
			deviceID,
			cfg.TCP.FrameLength,
			cfg.TCP.ReadTimeout,
			cfg.TCP.IdleTimeout,
			dataHandler,
		)
		serverManager.AddServer(server)
	}

	// 启动所有TCP服务器
	if err := serverManager.StartAll(); err != nil {
		log.Fatalf("Start TCP servers error: %v", err)
	}

	log.Printf("=== TCP Collector Started ===")
	log.Printf("Listening on ports: %v", cfg.TCP.Ports)
	log.Printf("Kafka topic: %s", cfg.Kafka.Topic)
	log.Printf("Press Ctrl+C to stop...")

	// 等待信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	log.Printf("Received signal: %v, shutting down gracefully...", sig)

	// 优雅关闭
	gracefulShutdown(serverManager, kafkaProducer)

	log.Printf("=== TCP Collector Stopped ===")
}

// gracefulShutdown 优雅关闭
func gracefulShutdown(serverManager *tcpserver.Manager, kafkaProducer *kafka.Producer) {
	log.Printf("Starting graceful shutdown...")

	// 1. 停止TCP服务器(不再接受新连接)
	if err := serverManager.StopAll(); err != nil {
		log.Printf("Stop TCP servers error: %v", err)
	}

	// 2. 等待一段时间,让正在处理的数据完成
	time.Sleep(2 * time.Second)

	// 3. 关闭Kafka生产者(发送剩余数据)
	if err := kafkaProducer.Close(); err != nil {
		log.Printf("Close kafka producer error: %v", err)
	}

	log.Printf("Graceful shutdown completed")
}
