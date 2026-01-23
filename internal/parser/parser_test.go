package parser

import (
	"encoding/hex"
	"testing"
)

func TestParseFloat32(t *testing.T) {
	p := NewParser()

	// 测试用例: 44 29 D4 DA 应该解析为约679.321
	hexStr := "4429D4DA"
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		t.Fatalf("decode hex error: %v", err)
	}

	result, err := p.parseFloat32(data)
	if err != nil {
		t.Fatalf("parse float32 error: %v", err)
	}

	// 验证结果在合理范围内(679.3左右)
	expected := 679.321
	diff := result - expected
	if diff < -1 || diff > 1 {
		t.Errorf("parse float32 result mismatch: expected ~%f, got %f", expected, result)
	}

	t.Logf("Parse result: %f", result)
}

func TestParseFrame(t *testing.T) {
	p := NewParser()

	// 构造一个71字节的测试数据
	data := make([]byte, 71)

	// PT1: 44 29 D4 DA (679.321)
	pt1Bytes, _ := hex.DecodeString("4429D4DA")
	copy(data[0:4], pt1Bytes)

	// PT2: 43 F0 00 00 (480.0)
	pt2Bytes, _ := hex.DecodeString("43F00000")
	copy(data[4:8], pt2Bytes)

	// 填充其他字节为0
	for i := 8; i < 71; i++ {
		data[i] = 0x00
	}

	// byte56 LS11: 1
	data[56] = 0x01
	// byte57 LS21: 2
	data[57] = 0x02

	// 解析
	result, err := p.Parse(data, "TEST-001", 1737654321000)
	if err != nil {
		t.Fatalf("parse frame error: %v", err)
	}

	// 验证结果
	if result.DeviceID != "TEST-001" {
		t.Errorf("device_id mismatch: expected TEST-001, got %s", result.DeviceID)
	}

	if result.Timestamp != 1737654321000 {
		t.Errorf("timestamp mismatch: expected 1737654321000, got %d", result.Timestamp)
	}

	// 验证PT1值
	if len(result.Data) < 1 {
		t.Fatal("no data points")
	}

	pt1 := result.Data[0]
	if pt1.Quota != "PT1" {
		t.Errorf("PT1 quota mismatch: expected PT1, got %s", pt1.Quota)
	}

	// PT1应该约为679.321
	if pt1.Value < 678 || pt1.Value > 680 {
		t.Errorf("PT1 value out of range: %f", pt1.Value)
	}

	t.Logf("Parse result: %+v", result)
	t.Logf("PT1: %+v", pt1)
}
