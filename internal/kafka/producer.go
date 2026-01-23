package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"tcp-collector/internal/config"
	"tcp-collector/internal/parser"
)

// Producer Kafka生产者
type Producer struct {
	producer     sarama.SyncProducer
	topic        string
	batchEnabled bool
	batchSize    int
	batchTimeout time.Duration
	batch        []*parser.ParsedData
	batchMu      sync.Mutex
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

// NewProducer 创建新的Kafka生产者
func NewProducer(cfg *config.KafkaConfig) (*Producer, error) {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Retry.Max = cfg.MaxRetries
	kafkaConfig.Producer.Return.Successes = true

	// 设置压缩方式
	switch cfg.Compression {
	case "gzip":
		kafkaConfig.Producer.Compression = sarama.CompressionGZIP
	case "snappy":
		kafkaConfig.Producer.Compression = sarama.CompressionSnappy
	case "lz4":
		kafkaConfig.Producer.Compression = sarama.CompressionLZ4
	case "zstd":
		kafkaConfig.Producer.Compression = sarama.CompressionZSTD
	default:
		kafkaConfig.Producer.Compression = sarama.CompressionNone
	}

	producer, err := sarama.NewSyncProducer(cfg.Brokers, kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("create kafka producer error: %w", err)
	}

	p := &Producer{
		producer:     producer,
		topic:        cfg.Topic,
		batchEnabled: cfg.BatchEnabled,
		batchSize:    cfg.BatchSize,
		batchTimeout: time.Duration(cfg.BatchTimeout) * time.Millisecond,
		batch:        make([]*parser.ParsedData, 0, cfg.BatchSize),
		stopCh:       make(chan struct{}),
	}

	// 如果启用批量发送,启动批量发送协程
	if p.batchEnabled {
		p.wg.Add(1)
		go p.batchSendLoop()
	}

	log.Printf("[Kafka] Producer created, brokers: %v, topic: %s", cfg.Brokers, cfg.Topic)
	return p, nil
}

// Send 发送数据
func (p *Producer) Send(data *parser.ParsedData) error {
	if p.batchEnabled {
		return p.addToBatch(data)
	}

	return p.sendImmediately(data)
}

// sendImmediately 立即发送
func (p *Producer) sendImmediately(data *parser.ParsedData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal data error: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(jsonData),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("send message to kafka error: %w", err)
	}

	log.Printf("[Kafka] Message sent: partition=%d, offset=%d, device=%s", partition, offset, data.DeviceID)
	return nil
}

// addToBatch 添加到批量发送队列
func (p *Producer) addToBatch(data *parser.ParsedData) error {
	p.batchMu.Lock()
	defer p.batchMu.Unlock()

	p.batch = append(p.batch, data)

	// 如果达到批量大小,立即发送
	if len(p.batch) >= p.batchSize {
		return p.flushBatchLocked()
	}

	return nil
}

// flushBatchLocked 发送批量数据(已加锁)
func (p *Producer) flushBatchLocked() error {
	if len(p.batch) == 0 {
		return nil
	}

	messages := make([]*sarama.ProducerMessage, 0, len(p.batch))
	for _, data := range p.batch {
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Printf("[Kafka] Marshal data error: %v", err)
			continue
		}

		messages = append(messages, &sarama.ProducerMessage{
			Topic: p.topic,
			Value: sarama.ByteEncoder(jsonData),
		})
	}

	if len(messages) == 0 {
		p.batch = p.batch[:0]
		return nil
	}

	// 批量发送
	err := p.producer.SendMessages(messages)
	if err != nil {
		// 部分失败的情况
		if errs, ok := err.(sarama.ProducerErrors); ok {
			log.Printf("[Kafka] Batch send partial failure: %d/%d messages failed", len(errs), len(messages))
			for _, e := range errs {
				log.Printf("[Kafka] Message error: partition=%d, error=%v", e.Msg.Partition, e.Err)
			}
		} else {
			return fmt.Errorf("send messages to kafka error: %w", err)
		}
	}

	log.Printf("[Kafka] Batch sent: %d messages", len(messages))
	p.batch = p.batch[:0]
	return nil
}

// batchSendLoop 批量发送循环
func (p *Producer) batchSendLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.batchTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 定时发送
			p.batchMu.Lock()
			if err := p.flushBatchLocked(); err != nil {
				log.Printf("[Kafka] Flush batch error: %v", err)
			}
			p.batchMu.Unlock()

		case <-p.stopCh:
			// 停止前发送剩余数据
			p.batchMu.Lock()
			if err := p.flushBatchLocked(); err != nil {
				log.Printf("[Kafka] Final flush batch error: %v", err)
			}
			p.batchMu.Unlock()
			return
		}
	}
}

// Close 关闭生产者
func (p *Producer) Close() error {
	log.Printf("[Kafka] Closing producer")

	if p.batchEnabled {
		close(p.stopCh)
		p.wg.Wait()
	}

	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("close kafka producer error: %w", err)
	}

	log.Printf("[Kafka] Producer closed")
	return nil
}
