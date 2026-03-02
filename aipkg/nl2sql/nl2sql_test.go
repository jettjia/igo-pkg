package nl2sql

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/jettjia/igo-pkg/aipkg/models/chat"
	"github.com/jettjia/igo-pkg/aipkg/pkg/types"
	"github.com/jettjia/igo-pkg/aipkg/schema"
	config "github.com/jettjia/igo-pkg/pkg/conf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/assert/yaml"
)

// 定义提示词结构体
type Prompts struct {
	NL2SQLPromptSystem string `yaml:"nl2sql_prompt_system" json:"nl2sql_prompt_system"`
	NL2SQLPromptUser   string `yaml:"nl2sql_prompt_user" json:"nl2sql_prompt_user"`
}

// 实现getPrompt函数，从YAML文件读取提示词
func getPrompt() (*Prompts, error) {
	// 获取当前文件所在目录
	currentDir, err := filepath.Abs(filepath.Dir(""))
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

// 测试基本SQL生成功能
// go test -v -run Test_NL2SQL_GenerateSQL ./
func Test_NL2SQL_GenerateSQL(t *testing.T) {
	// 初始化配置
	conf := &config.Config{}
	prompts, err := getPrompt()
	if err != nil {
		panic(err)
	}
	conf.Ai.NL2SQLPromptSystem = prompts.NL2SQLPromptSystem
	conf.Ai.NL2SQLPromptUser = prompts.NL2SQLPromptUser

	// 获取LLM客户端
	llm := getLLM()

	// 创建NL2SQL实例
	nl2sql, err := NewNL2SQL(llm, conf)
	if err != nil {
		panic(err)
	}

	// 定义测试用例
	testCases := []struct {
		name          string
		query         string
		tableSchema   []*schema.TableSchema
		expectedError bool
	}{{
		name:  "简单查询测试",
		query: "查询所有用户的姓名和邮箱",
		tableSchema: []*schema.TableSchema{{
			Name: "users",
			Columns: []*schema.Column{{
				Name: "id",
				Type: schema.TableColumnTypeInteger,
			}, {
				Name: "name",
				Type: schema.TableColumnTypeString,
			}, {
				Name: "email",
				Type: schema.TableColumnTypeString,
			}},
		}},
		expectedError: false,
	}, {
		name:  "条件查询测试",
		query: "查询年龄大于18的用户",
		tableSchema: []*schema.TableSchema{{
			Name: "users",
			Columns: []*schema.Column{{
				Name: "id",
				Type: schema.TableColumnTypeInteger,
			}, {
				Name: "name",
				Type: schema.TableColumnTypeString,
			}, {
				Name: "age",
				Type: schema.TableColumnTypeInteger,
			}},
		}},
		expectedError: false,
	}}

	// 执行测试用例
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 构建消息
			messages := []*types.Message{{
				Role:    "user",
				Content: tc.query,
			}}

			// 生成SQL
			sql, err := nl2sql.NL2SQL(ctx, messages, tc.tableSchema)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, sql)
				t.Logf("原始问题: %s", tc.query)
				t.Logf("生成的SQL: %s", sql)
			}
		})
	}
}
