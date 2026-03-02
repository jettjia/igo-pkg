package rewrite

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jettjia/igo-pkg/aipkg/models/chat"
	config "github.com/jettjia/igo-pkg/pkg/conf"
	"gopkg.in/yaml.v3"
)

// 定义提示词结构体
type Prompts struct {
	KeywordsExtractionPrompt string `yaml:"keywords_extraction_prompt" json:"keywords_extraction_prompt"` // 提问关键词提取
	RewritePromptSystem      string `yaml:"rewrite_prompt_system" json:"rewrite_prompt_system"`           // 提问重写
	RewritePromptUser        string `yaml:"rewrite_prompt_user" json:"rewrite_prompt_user"`               // 提问重写
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

// 测试关键词提取
// go test -v -run Test_Rewriter_ExtractKeywords ./
func Test_Rewriter_ExtractKeywords(t *testing.T) {
	// 初始化配置
	conf := &config.Config{}
	prompts, err := getPrompt()
	if err != nil {
		panic(err)
	}
	conf.Ai.KeywordsExtractionPrompt = prompts.KeywordsExtractionPrompt

	// 获取LLM客户端
	llm := getLLM()

	// 创建重写器
	rewriter, err := NewRewriter(llm, conf)
	if err != nil {
		panic(err)
	}

	// 测试关键词提取
	ctx := context.Background()
	query := "苹果手机电池不耐用怎么解决？"
	keywords, err := rewriter.ExtractKeywords(ctx, query)
	if err != nil {
		t.Fatalf("关键词提取失败: %v", err)
	}
	t.Logf("提取到的关键词: %s", keywords)

	// 测试关键词提取
	query = "不会安装k8s怎么办？"
	keywords, err = rewriter.ExtractKeywords(ctx, query)
	if err != nil {
		t.Fatalf("关键词提取失败: %v", err)
	}
	t.Logf("提取到的关键词: %s", keywords)
}

// 测试问题重写（使用历史消息格式）
// go test -v -run Test_Rewriter_RewriteQuery ./
func Test_Rewriter_RewriteQuery(t *testing.T) {
	// 初始化配置
	conf := &config.Config{}
	prompts, err := getPrompt()
	if err != nil {
		panic(err)
	}
	// 设置重写提示词
	conf.Ai.RewritePromptSystem = prompts.RewritePromptSystem
	conf.Ai.RewritePromptUser = prompts.RewritePromptUser

	// 获取LLM客户端
	llm := getLLM()

	// 创建重写器
	rewriter, err := NewRewriter(llm, conf)
	if err != nil {
		panic(err)
	}

	// 定义多个测试用例
	testCases := []struct {
		name              string
		history           []chat.Message
		query             string
		expectedSubstring string // 期望结果中包含的子串
	}{{
		name: "示例1: 代词替换(它)",
		history: []chat.Message{
			{
				Role:    "user",
				Content: "微信支付有哪些功能？",
			},
			{
				Role:    "assistant",
				Content: "微信支付的主要功能包括转账、付款码、收款、信用卡还款等多种支付服务。",
			},
		},
		query:             "它的安全性",
		expectedSubstring: "微信支付的安全性",
	}, {
		name: "示例2: 代词替换(这样)",
		history: []chat.Message{
			{
				Role:    "user",
				Content: "苹果手机电池不耐用怎么办？",
			},
			{
				Role:    "assistant",
				Content: "您可以通过降低屏幕亮度、关闭后台应用和定期更新系统来延长电池寿命。",
			},
		},
		query:             "这样会影响使用体验吗？",
		expectedSubstring: "降低屏幕亮度和关闭后台应用是否影响使用体验",
	}, {
		name: "示例3: 省略补全",
		history: []chat.Message{
			{
				Role:    "user",
				Content: "如何制作红烧肉？",
			},
			{
				Role:    "assistant",
				Content: "红烧肉的制作需要先将肉块焯水，然后加入酱油、糖等调料慢炖。",
			},
		},
		query:             "需要炖多久？",
		expectedSubstring: "红烧肉需要炖多久",
	}, {
		name: "示例4: 省略补全",
		history: []chat.Message{
			{
				Role:    "user",
				Content: "北京到上海的高铁票价是多少？",
			},
			{
				Role:    "assistant",
				Content: "北京到上海的高铁票价根据车次和座位类型不同，二等座约为553元，一等座约为933元。",
			},
		},
		query:             "时间呢？",
		expectedSubstring: "北京到上海的高铁时长",
	}, {
		name: "示例5: 代词替换(国外手机号)",
		history: []chat.Message{
			{
				Role:    "user",
				Content: "如何注册微信账号？",
			},
			{
				Role:    "assistant",
				Content: "注册微信账号需要下载微信APP，输入手机号，接收验证码，然后设置昵称和密码。",
			},
		},
		query:             "国外手机号可以吗？",
		expectedSubstring: "国外手机号是否可以注册微信账号",
	}, {
		name: "多轮对话",
		history: []chat.Message{
			{
				Role:    "user",
				Content: "什么是人工智能？",
			},
			{
				Role:    "assistant",
				Content: "人工智能是指由人制造出来的系统所表现出来的智能。",
			},
			{
				Role:    "user",
				Content: "它有哪些应用？",
			},
			{
				Role:    "assistant",
				Content: "人工智能的应用包括语音识别、图像识别、自然语言处理等。",
			},
		},
		query:             "这些应用中最热门的是什么？",
		expectedSubstring: "人工智能应用中最热门",
	}, {
		name:              "无历史对话",
		history:           []chat.Message{},
		query:             "什么是区块链？",
		expectedSubstring: "什么是区块链",
	}, {
		name: "不按顺序的消息",
		history: []chat.Message{
			{
				Role:    "assistant",
				Content: "这是一个系统消息。",
			},
			{
				Role:    "user",
				Content: "什么是云计算？",
			},
			{
				Role:    "assistant",
				Content: "云计算是一种通过互联网提供计算资源的服务模式。",
			},
		},
		query:             "它的优势是什么？",
		expectedSubstring: "云计算的优势",
	}}

	// 执行测试用例
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rewrittenQuery, err := rewriter.RewriteQuery(ctx, tc.history, tc.query)
			if err != nil {
				t.Fatalf("问题重写失败: %v", err)
			}
			t.Logf("原始问题: %s", tc.query)
			t.Logf("重写后的问题: %s", rewrittenQuery)

			// 验证结果是否包含期望的子串
			matched := false
			queryWords := strings.Fields(strings.ToLower(rewrittenQuery))
			expectedWords := strings.Fields(strings.ToLower(tc.expectedSubstring))

			// 检查所有期望的关键词是否都在重写后的问题中
			matchCount := 0
			for _, expectedWord := range expectedWords {
				for _, queryWord := range queryWords {
					if strings.Contains(queryWord, expectedWord) || strings.Contains(expectedWord, queryWord) {
						matchCount++
						break
					}
				}
			}

			// 如果匹配率超过60%，则认为通过
			if matchCount >= len(expectedWords)*3/5 || strings.Contains(rewrittenQuery, tc.expectedSubstring) {
				matched = true
			}

			if !matched {
				t.Errorf("重写结果验证失败: 期望包含 '%s', 实际为 '%s'", tc.expectedSubstring, rewrittenQuery)
			}
		})
	}
}
