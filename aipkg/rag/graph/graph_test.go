package graph

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jettjia/go-pkg/aipkg/models/chat"
	"github.com/jettjia/go-pkg/aipkg/pkg/types"
	config "github.com/jettjia/go-pkg/pkg/conf"
	"github.com/jettjia/go-pkg/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/assert/yaml"
)

// 定义提示词结构体
type Prompts struct {
	ExtractEntitiesPrompt      string `yaml:"extract_entities_prompt"`
	ExtractRelationshipsPrompt string `yaml:"extract_relationships_prompt"`
}

// 实现getPrompt函数，从YAML文件读取提示词
func getPrompt() (*Prompts, error) {
	// 获取当前文件所在目录
	currentDir, err := filepath.Abs(filepath.Dir("."))
	if err != nil {
		return nil, fmt.Errorf("获取当前目录失败: %w", err)
	}

	// 构建YAML文件路径
	promptFilePath := filepath.Join(currentDir, "prompt.example.yaml")

	// 读取YAML文件内容
	data, err := ioutil.ReadFile(promptFilePath)
	if err != nil {
		return nil, fmt.Errorf("读取提示词文件失败: %w", err)
	}

	// 解析YAML
	var prompts Prompts
	if err := yaml.Unmarshal(data, &prompts); err != nil {
		return nil, fmt.Errorf("解析提示词文件失败: %w", err)
	}

	return &prompts, nil
}

func getLLM() chat.Chat {
	chatModel, err := chat.NewChat(&chat.ChatConfig{
		ModelID:   "1",
		APIKey:    os.Getenv("OPENAI_API_KEY"),
		BaseURL:   os.Getenv("OPENAI_BASE_URL"),
		ModelName: "deepseek-ai/DeepSeek-V3",
		Source:    "remote",
	})
	if err != nil {
		panic(err)
	}

	return chatModel
}

// Test_GraphBuilder_BuildGraph 测试图谱构建
// go test -v -run Test_GraphBuilder_BuildGraph ./
func Test_GraphBuilder_BuildGraph(t *testing.T) {
	// 初始化配置
	conf := &config.Config{}
	prompts, err := getPrompt()
	if err != nil {
		panic(err)
	}
	conf.Ai.ExtractEntitiesPrompt = prompts.ExtractEntitiesPrompt
	conf.Ai.ExtractRelationshipsPrompt = prompts.ExtractRelationshipsPrompt

	// 获取LLM客户端
	llm := getLLM()

	// 创建图谱构建器
	graphBuilder := NewGraphBuilder(conf, llm).(*graphBuilder)

	// 创建测试用的文档块
	chunks := []*types.Chunk{
		{
			ID:      uuid.New().String(),
			Content: "DeepSeek-R1是硅基流动开发的大语言模型，可用于构建知识图谱。",
			StartAt: 0,
			EndAt:   len("DeepSeek-R1是硅基流动开发的大语言模型，可用于构建知识图谱。"),
		},
		{
			ID:      uuid.New().String(),
			Content: "知识图谱是一种用于表示实体及其关系的数据结构，DeepSeek-R1可以利用知识图谱提升理解能力。",
			StartAt: 0,
			EndAt:   len("知识图谱是一种用于表示实体及其关系的数据结构，DeepSeek-R1可以利用知识图谱提升理解能力。"),
		},
	}

	// 构建图谱
	ctx := context.Background()
	err = graphBuilder.BuildGraph(ctx, chunks)
	assert.NoError(t, err)

	// 验证实体提取结果
	assert.Greater(t, len(graphBuilder.entityMap), 0, "至少应提取到一个实体")
	assert.Contains(t, graphBuilder.entityMapByTitle, "DeepSeek-R1", "应包含'DeepSeek-R1'实体")
	assert.Contains(t, graphBuilder.entityMapByTitle, "知识图谱", "应包含'知识图谱'实体")

	// 验证关系提取结果
	deepSeekEntity := graphBuilder.entityMapByTitle["DeepSeek-R1"]
	knowledgeGraphEntity := graphBuilder.entityMapByTitle["知识图谱"]
	relationKey := deepSeekEntity.Title + "#" + knowledgeGraphEntity.Title
	assert.Contains(t, graphBuilder.relationshipMap, relationKey, "应包含DeepSeek-R1到知识图谱的关系")

	// 验证关系属性
	relation := graphBuilder.relationshipMap[relationKey]
	assert.Equal(t, deepSeekEntity.Title, relation.Source, "关系源实体应为DeepSeek-R1")
	assert.Equal(t, knowledgeGraphEntity.Title, relation.Target, "关系目标实体应为知识图谱")
	assert.Greater(t, relation.Strength, 0, "关系强度应大于0")
	assert.Greater(t, len(relation.ChunkIDs), 0, "关系应关联至少一个文档块")

	// 验证块图构建
	assert.Greater(t, len(graphBuilder.chunkGraph), 0, "块图应非空")

	// 打印
	fmt.Println(util.PrintJson(graphBuilder.chunkGraph))
}

// 测试实体提取功能
func Test_GraphBuilder_ExtractEntities(t *testing.T) {
	// 初始化配置
	conf := &config.Config{}
	prompts, err := getPrompt()
	if err != nil {
		panic(err)
	}
	conf.Ai.ExtractEntitiesPrompt = prompts.ExtractEntitiesPrompt

	// 获取LLM客户端
	llm := getLLM()

	// 创建图谱构建器
	graphBuilder := NewGraphBuilder(conf, llm).(*graphBuilder)

	// 创建测试文档块
	chunk := &types.Chunk{
		ID:      uuid.New().String(),
		Content: "DeepSeek-R1是硅基流动开发的大语言模型，可用于构建知识图谱。",
		StartAt: 0,
		EndAt:   len("DeepSeek-R1是硅基流动开发的大语言模型，可用于构建知识图谱。"),
	}

	// 提取实体
	ctx := context.Background()
	entities, err := graphBuilder.extractEntities(ctx, chunk)
	assert.NoError(t, err)

	// 验证结果
	assert.Len(t, entities, 2, "应提取到2个实体")
	assert.Equal(t, "DeepSeek-R1", entities[0].Title, "第一个实体应为DeepSeek-R1")
	assert.Equal(t, "知识图谱", entities[1].Title, "第二个实体应为知识图谱")
	assert.Contains(t, entities[0].ChunkIDs, chunk.ID, "实体应关联到文档块")

	// 打印
	fmt.Println(entities)
}

// 测试关系提取功能
func Test_GraphBuilder_ExtractRelationships(t *testing.T) {
	// 初始化配置
	conf := &config.Config{}
	prompts, err := getPrompt()
	if err != nil {
		panic(err)
	}
	conf.Ai.ExtractRelationshipsPrompt = prompts.ExtractRelationshipsPrompt

	// 获取LLM客户端
	llm := getLLM()

	// 创建图谱构建器
	graphBuilder := NewGraphBuilder(conf, llm).(*graphBuilder)

	// 创建测试文档块和实体
	chunk := &types.Chunk{
		ID:      uuid.New().String(),
		Content: "DeepSeek-R1是硅基流动开发的大语言模型，可用于构建知识图谱。",
		StartAt: 0,
		EndAt:   len("DeepSeek-R1是硅基流动开发的大语言模型，可用于构建知识图谱。"),
	}

	entities := []*types.Entity{
		{
			ID:          uuid.New().String(),
			Title:       "DeepSeek-R1",
			Description: "硅基流动开发的大语言模型",
			ChunkIDs:    []string{chunk.ID},
		},
		{
			ID:          uuid.New().String(),
			Title:       "知识图谱",
			Description: "用于表示实体及其关系的数据结构",
			ChunkIDs:    []string{chunk.ID},
		},
	}

	// 提取关系
	ctx := context.Background()
	// 提取关系（带重试）
	maxRetries := 3
	var retryCount int
	for retryCount < maxRetries {
		err = graphBuilder.extractRelationships(ctx, []*types.Chunk{chunk}, entities)
		if err == nil {
			break // 成功，退出循环
		}

		retryCount++
		if retryCount < maxRetries && strings.Contains(err.Error(), "invalid character '`'") {
			t.Logf("第%d次尝试失败，检测到无效字符'`'，%d秒后重试...", retryCount, retryCount*2)
			time.Sleep(time.Duration(retryCount*2) * time.Second)
		} else {
			t.Fatalf("提取关系失败: %v", err)
		}
	}
	assert.NoError(t, err)

	// 验证结果
	relationKey := "DeepSeek-R1#知识图谱"
	assert.Contains(t, graphBuilder.relationshipMap, relationKey, "应提取到DeepSeek-R1到知识图谱的关系")

	relation := graphBuilder.relationshipMap[relationKey]
	assert.Equal(t, "DeepSeek-R1", relation.Source, "关系源实体应为DeepSeek-R1")
	assert.Equal(t, "知识图谱", relation.Target, "关系目标实体应为知识图谱")
	assert.Equal(t, "用于构建", relation.Description, "关系描述应为'用于构建'")
	assert.GreaterOrEqual(t, relation.Strength, 5, "关系强度应大于等于5")

	// 打印
	fmt.Println(relation)
}

// 测试关系提取功能 (使用大模型提取实体和关系)
// go test -v -run Test_GraphBuilder_ExtSanguo ./
func Test_GraphBuilder_ExtSanguo(t *testing.T) {
	// 记录开始时间
	startTime := time.Now()
	defer func() {
		// 记录结束时间并计算耗时
		elapsedTime := time.Since(startTime)
		fmt.Printf("测试执行耗时: %v\n", elapsedTime)
	}()
	// 读取CSV文件数据
	triples, err := readTriplesFromCSV("triples.csv")
	if err != nil {
		t.Fatalf("读取triples.csv失败: %v", err)
	}

	// 构建测试文本
	var testText strings.Builder
	for _, triple := range triples {
		testText.WriteString(fmt.Sprintf("%s是%s的%s（%s）。", triple.Head, triple.Tail, triple.Label, triple.Relation))
	}

	// 获取LLM客户端
	llm := getLLM()

	// 初始化配置
	conf := &config.Config{}
	prompts, err := getPrompt()
	if err != nil {
		panic(err)
	}
	conf.Ai.ExtractRelationshipsPrompt = prompts.ExtractRelationshipsPrompt
	conf.Ai.ExtractEntitiesPrompt = prompts.ExtractEntitiesPrompt

	// 创建GraphBuilder
	graphBuilder := NewGraphBuilder(conf, llm).(*graphBuilder)

	// 使用大模型提取实体
	chunk := &types.Chunk{
		ID:      uuid.New().String(),
		Content: testText.String(),
	}
	ctx := context.Background()
	entities, err := graphBuilder.extractEntities(ctx, chunk)
	if err != nil {
		t.Fatalf("提取实体失败: %v", err)
	}
	fmt.Println("==============print entities")
	fmt.Println(util.PrintJson(entities))

	// 提取关系
	err = graphBuilder.extractRelationships(ctx, []*types.Chunk{chunk}, entities)
	if err != nil {
		t.Fatalf("提取关系失败: %v", err)
	}
	fmt.Println("==============print relationships")
	fmt.Println(util.PrintJson(graphBuilder.relationshipMap))
}

// Triple 定义CSV中的三元组数据结构
type Triple struct {
	Head     string
	Tail     string
	Relation string
	Label    string
}

// readTriplesFromCSV 从CSV文件读取三元组数据
func readTriplesFromCSV(filePath string) ([]Triple, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// 跳过表头
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("读取表头失败: %w", err)
	}

	var triples []Triple
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取记录失败: %w", err)
		}

		// 确保记录有4个字段
		if len(record) < 4 {
			continue
		}

		triples = append(triples, Triple{
			Head:     record[0],
			Tail:     record[1],
			Relation: record[2],
			Label:    record[3],
		})
	}

	return triples, nil
}
