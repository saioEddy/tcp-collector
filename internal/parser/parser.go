package parser

import (
	"encoding/binary"
	"fmt"
	"math"
)

// DataPoint 单个数据点
type DataPoint struct {
	Quota string  `json:"quota"` // 指标名称
	Value float64 `json:"value"` // 指标值
}

// ParsedData 解析后的数据结构
type ParsedData struct {
	Timestamp int64       `json:"timestamp"` // 时间戳(毫秒)
	DeviceID  string      `json:"device_id"` // 设备ID
	Data      []DataPoint `json:"data"`      // 数据点列表
}

// Parser 数据解析器
type Parser struct {
	fields []FieldDef
}

// NewParser 创建新的解析器
func NewParser() *Parser {
	return &Parser{
		fields: ParseStructFields,
	}
}

// Parse 解析HEX数据
// data: 原始字节数据(71字节)
// deviceID: 设备ID
// timestamp: 时间戳(毫秒)
func (p *Parser) Parse(data []byte, deviceID string, timestamp int64) (*ParsedData, error) {
	if len(data) != FrameLength {
		return nil, fmt.Errorf("invalid data length: expected %d, got %d", FrameLength, len(data))
	}

	parsed := &ParsedData{
		Timestamp: timestamp,
		DeviceID:  deviceID,
		Data:      make([]DataPoint, 0, len(p.fields)),
	}

	for _, field := range p.fields {
		value, err := p.parseField(data, field)
		if err != nil {
			return nil, fmt.Errorf("parse field %s error: %w", field.Name, err)
		}

		parsed.Data = append(parsed.Data, DataPoint{
			Quota: field.Name,
			Value: value,
		})
	}

	return parsed, nil
}

// parseField 解析单个字段
func (p *Parser) parseField(data []byte, field FieldDef) (float64, error) {
	switch field.Type {
	case FieldTypeFloat32:
		return p.parseFloat32(data[field.StartByte : field.EndByte+1])
	case FieldTypeUint8:
		return float64(data[field.StartByte]), nil
	case FieldTypeUint16:
		return p.parseUint16(data[field.StartByte : field.EndByte+1])
	default:
		return 0, fmt.Errorf("unknown field type: %d", field.Type)
	}
}

// parseFloat32 解析IEEE-754 32位浮点数(大端序)
func (p *Parser) parseFloat32(data []byte) (float64, error) {
	if len(data) != 4 {
		return 0, fmt.Errorf("invalid float32 data length: %d", len(data))
	}

	// 大端序读取32位整数
	bits := binary.BigEndian.Uint32(data)
	// 转换为float32
	floatVal := math.Float32frombits(bits)
	// 转换为float64返回
	return float64(floatVal), nil
}

// parseUint16 解析16位无符号整数(大端序)
func (p *Parser) parseUint16(data []byte) (float64, error) {
	if len(data) != 2 {
		return 0, fmt.Errorf("invalid uint16 data length: %d", len(data))
	}

	// 大端序读取16位整数
	value := binary.BigEndian.Uint16(data)
	return float64(value), nil
}
