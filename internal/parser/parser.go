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
		// 预留一点空间：字段解析 + 可能的合成字段（如 FS1）
		Data: make([]DataPoint, 0, len(p.fields)+4),
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

	// 合成字段（已停用）：
	// 你提供的“解析结构文件”里 FS1 明确拆成 FS1-C1/FS1-C2，
	// 下游如果把 FS1 也当成独立指标，会产生重复字段/歧义，因此默认不再额外输出 FS1。
	//
	// 如确实需要兼容旧消费者，可恢复下面这段代码：
	// fs1 := binary.BigEndian.Uint16(data[62:64]) // FS1 = C1(高) + C2(低)
	// parsed.Data = append(parsed.Data, DataPoint{Quota: "FS1", Value: float64(fs1)})

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
	case FieldTypeUint48:
		return p.parseUint48(data[field.StartByte : field.EndByte+1])
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

// parseUint48 解析48位无符号整数(大端序)
// 注意：返回 float64 不会丢精度，因为 uint48 < 2^53。
func (p *Parser) parseUint48(data []byte) (float64, error) {
	if len(data) != 6 {
		return 0, fmt.Errorf("invalid uint48 data length: %d", len(data))
	}

	var v uint64
	for i := 0; i < 6; i++ {
		v = (v << 8) | uint64(data[i])
	}
	return float64(v), nil
}
