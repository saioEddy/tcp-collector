package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	TCP    TCPConfig    `yaml:"tcp"`
	Kafka  KafkaConfig  `yaml:"kafka"`
	Log    LogConfig    `yaml:"log"`
	Parser ParserConfig `yaml:"parser"`
}

// TCPConfig TCP配置
type TCPConfig struct {
	Ports       []int    `yaml:"ports"`
	DeviceIDs   []string `yaml:"device_ids"`
	FrameLength int      `yaml:"frame_length"`
	ReadTimeout int      `yaml:"read_timeout"`
	IdleTimeout int      `yaml:"idle_timeout"`
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers      []string `yaml:"brokers"`
	Topic        string   `yaml:"topic"`
	BatchEnabled bool     `yaml:"batch_enabled"`
	BatchSize    int      `yaml:"batch_size"`
	BatchTimeout int      `yaml:"batch_timeout"`
	Compression  string   `yaml:"compression"`
	MaxRetries   int      `yaml:"max_retries"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string `yaml:"level"`
	Output   string `yaml:"output"`
	FilePath string `yaml:"file_path"`
	MaxSize  int    `yaml:"max_size"`
	MaxDays  int    `yaml:"max_days"`
}

// ParserConfig 解析器配置
type ParserConfig struct {
	StructFile string `yaml:"struct_file"`
}

// LoadConfig 加载配置文件
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file error: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config error: %w", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config error: %w", err)
	}

	return &cfg, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	if len(c.TCP.Ports) == 0 {
		return fmt.Errorf("tcp.ports is empty")
	}

	if len(c.TCP.DeviceIDs) != len(c.TCP.Ports) {
		return fmt.Errorf("tcp.device_ids length must equal tcp.ports length")
	}

	if c.TCP.FrameLength <= 0 {
		return fmt.Errorf("tcp.frame_length must be positive")
	}

	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("kafka.brokers is empty")
	}

	if c.Kafka.Topic == "" {
		return fmt.Errorf("kafka.topic is empty")
	}

	return nil
}
