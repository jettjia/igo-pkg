package chunker

// StrategyType 定义分块策略类型的自定义类型
type StrategyType string

// 记忆类型常量定义
const (
	// StrategyTypeRecursiveCharacter 递归字符分块策略
	StrategyTypeRecursiveCharacter StrategyType = "recursive_character"

	// StrategyTypeRecursiveParagraph 递归段落分块策略
	StrategyTypeRecursiveParagraph StrategyType = "recursive_paragraph"

	// StrategyTypeSemantic 语义分块策略
	StrategyTypeSemantic StrategyType = "semantic"
)

// String 返回StrategyType的字符串表示
func (m StrategyType) String() string {
	return string(m)
}

// IsValid 验证StrategyType是否为有效的值
func (m StrategyType) IsValid() bool {
	switch m {
	case StrategyTypeRecursiveCharacter, StrategyTypeRecursiveParagraph, StrategyTypeSemantic:
		return true
	}
	return false
}

// GetAllStrategyTypes 获取所有有效的分块策略类型
func GetAllStrategyTypes() []StrategyType {
	return []StrategyType{
		StrategyTypeRecursiveCharacter,
		StrategyTypeRecursiveParagraph,
		StrategyTypeSemantic,
	}
}
