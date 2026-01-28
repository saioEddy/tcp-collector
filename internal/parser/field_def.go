package parser

// FieldType 字段类型
type FieldType int

const (
	// FieldTypeFloat32 IEEE-754 32位浮点数
	FieldTypeFloat32 FieldType = iota
	// FieldTypeUint8 8位无符号整数
	FieldTypeUint8
	// FieldTypeUint16 16位无符号整数
	FieldTypeUint16
)

// FieldDef 字段定义
type FieldDef struct {
	Name      string    // 字段名称,如"PT1"
	StartByte int       // 起始字节位置
	EndByte   int       // 结束字节位置
	Type      FieldType // 字段类型
}

// ParseStructFields 解析结构字段定义
// 根据parse_struct.txt文件生成的常量定义
var ParseStructFields = []FieldDef{
	// PT1-PT12: IEEE-754 32位浮点数,大端序
	{Name: "PT1", StartByte: 0, EndByte: 3, Type: FieldTypeFloat32},
	{Name: "PT2", StartByte: 4, EndByte: 7, Type: FieldTypeFloat32},
	{Name: "PT3", StartByte: 8, EndByte: 11, Type: FieldTypeFloat32},
	{Name: "PT4", StartByte: 12, EndByte: 15, Type: FieldTypeFloat32},
	{Name: "PT5", StartByte: 16, EndByte: 19, Type: FieldTypeFloat32},
	{Name: "PT6", StartByte: 20, EndByte: 23, Type: FieldTypeFloat32},
	{Name: "PT7", StartByte: 24, EndByte: 27, Type: FieldTypeFloat32},
	{Name: "PT8", StartByte: 28, EndByte: 31, Type: FieldTypeFloat32},
	{Name: "PT9", StartByte: 32, EndByte: 35, Type: FieldTypeFloat32},
	{Name: "PT10", StartByte: 36, EndByte: 39, Type: FieldTypeFloat32},
	{Name: "PT11", StartByte: 40, EndByte: 43, Type: FieldTypeFloat32},
	{Name: "PT12", StartByte: 44, EndByte: 47, Type: FieldTypeFloat32},

	// 温度和位移: IEEE-754 32位浮点数,大端序
	{Name: "温度", StartByte: 48, EndByte: 51, Type: FieldTypeFloat32},
	{Name: "位移", StartByte: 52, EndByte: 55, Type: FieldTypeFloat32},

	// LS系列: 8位无符号整数
	{Name: "LS11", StartByte: 56, EndByte: 56, Type: FieldTypeUint8},
	{Name: "LS21", StartByte: 57, EndByte: 57, Type: FieldTypeUint8},
	{Name: "LS31", StartByte: 58, EndByte: 58, Type: FieldTypeUint8},
	{Name: "LS11-1", StartByte: 59, EndByte: 59, Type: FieldTypeUint8},
	{Name: "LS21-1", StartByte: 60, EndByte: 60, Type: FieldTypeUint8},
	{Name: "LS31-1", StartByte: 61, EndByte: 61, Type: FieldTypeUint8},

	// FS1: 16位无符号整数,大端序(高字节+低字节)
	{Name: "FS1", StartByte: 62, EndByte: 63, Type: FieldTypeUint16},

	// 备用字段(64-69字节): 暂不解析,通常为保留字段或校验位
}

// FrameLength 数据帧长度(字节)
// 注意:实际有效数据到第63字节,总长度70字节(包含保留字段)
const FrameLength = 70
