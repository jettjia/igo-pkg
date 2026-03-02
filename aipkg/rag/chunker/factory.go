package chunker

import (
	"sync"
)

// StrategyFactory 策略工厂类，负责创建和管理各种分块策略
// 使用单例模式实现，确保全局只有一个工厂实例
// 支持策略注册、获取、列表查询和存在性检查
// 自动注册三种默认策略：递归字符分块、递归段落分块和语义分块

// StrategyFactory 策略工厂类
// 内部维护一个策略映射表，存储策略类型与策略创建函数的对应关系
// 使用读写锁保证并发安全
// - strategies: 存储策略类型名称到策略创建函数的映射
// - mutex: 保证并发安全的读写锁
type StrategyFactory struct {
	strategies map[StrategyType]func() ChunkingStrategy
	mutex      sync.RWMutex
}

var (
	factoryInstance *StrategyFactory
	factoryOnce     sync.Once
)

// GetStrategyFactory 获取工厂实例（单例模式）
// 确保全局只有一个工厂实例
// 在首次调用时会初始化工厂并自动注册默认策略
// 返回: 全局唯一的StrategyFactory实例
func GetStrategyFactory() *StrategyFactory {
	factoryOnce.Do(func() {
		factoryInstance = &StrategyFactory{
			strategies: make(map[StrategyType]func() ChunkingStrategy),
		}
		// 自动注册默认策略
		factoryInstance.registerDefaultStrategies()
	})
	return factoryInstance
}

// registerDefaultStrategies 注册默认策略
// 内部方法，在工厂初始化时自动调用
// 注册三种内置策略：
// 1. 递归字符分块策略（recursive_character）
// 2. 递归段落分块策略（recursive_paragraph）
// 3. 语义分块策略（semantic）
func (f *StrategyFactory) registerDefaultStrategies() {
	// 注册递归字符分块策略
	f.RegisterStrategy(StrategyTypeRecursiveCharacter, func() ChunkingStrategy {
		return NewRecursiveCharacterStrategy()
	})

	// 注册递归段落分块策略
	f.RegisterStrategy(StrategyTypeRecursiveParagraph, func() ChunkingStrategy {
		return NewRecursiveParagraphStrategy()
	})

	// 注册语义分块策略
	f.RegisterStrategy(StrategyTypeSemantic, func() ChunkingStrategy {
		return NewSemanticStrategy()
	})
}

// RegisterStrategy 注册一个新的分块策略
// 参数:
// - name: 策略名称，用于后续获取该策略
// - creator: 策略创建函数，返回一个新的策略实例
// 如果已存在同名策略，则会覆盖旧策略
// 线程安全，可以在并发环境下调用
func (f *StrategyFactory) RegisterStrategy(name StrategyType, creator func() ChunkingStrategy) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.strategies[name] = creator
}

// GetStrategy 根据名称获取一个策略实例
// 参数:
// - name: 策略名称
// 返回:
// - 第一个返回值: 策略实例，调用方需要自行处理返回的接口类型
// - 第二个返回值: 布尔值，表示是否找到对应策略
// 线程安全，可以在并发环境下调用
func (f *StrategyFactory) GetStrategy(name StrategyType) (ChunkingStrategy, bool) {
	f.mutex.RLock()
	creator, exists := f.strategies[name]
	f.mutex.RUnlock()

	if !exists {
		return nil, false
	}

	return creator(), true
}

// ListStrategies 获取所有已注册的策略名称
// 返回: 所有已注册策略名称的切片
// 线程安全，可以在并发环境下调用
func (f *StrategyFactory) ListStrategies() []StrategyType {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	strategies := make([]StrategyType, 0, len(f.strategies))
	for name := range f.strategies {
		strategies = append(strategies, name)
	}
	return strategies
}

// Exists 检查指定名称的策略是否已注册
// 参数:
// - name: 要检查的策略名称
// 返回: 布尔值，表示策略是否已注册
// 线程安全，可以在并发环境下调用
func (f *StrategyFactory) Exists(name StrategyType) bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	_, exists := f.strategies[name]
	return exists
}
