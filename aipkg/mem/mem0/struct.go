package mem0

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// APIVersion 定义 API 版本
type APIVersion string

const (
	V1 APIVersion = "v1"
	V2 APIVersion = "v2"
)

// Feedback 定义反馈类型
type Feedback string

const (
	Positive     Feedback = "POSITIVE"
	Negative     Feedback = "NEGATIVE"
	VeryNegative Feedback = "VERY_NEGATIVE"
)

// MemoryOptions 定义内存选项
type MemoryOptions struct {
	APIVersion         APIVersion       `json:"api_version,omitempty"`
	Version            APIVersion       `json:"version,omitempty"`
	UserID             string           `json:"user_id,omitempty"`
	AgentID            string           `json:"agent_id,omitempty"`
	AppID              string           `json:"app_id,omitempty"`
	RunID              string           `json:"run_id,omitempty"`
	Metadata           map[string]any   `json:"metadata,omitempty"`
	Filters            map[string]any   `json:"filters,omitempty"`
	OrgName            string           `json:"org_name,omitempty"`     // 已弃用
	ProjectName        string           `json:"project_name,omitempty"` // 已弃用
	OrgID              string           `json:"org_id,omitempty"`
	ProjectID          string           `json:"project_id,omitempty"`
	Infer              bool             `json:"infer,omitempty"`
	Page               int              `json:"page,omitempty"`
	PageSize           int              `json:"page_size,omitempty"`
	Includes           string           `json:"includes,omitempty"`
	Excludes           string           `json:"excludes,omitempty"`
	EnableGraph        bool             `json:"enable_graph,omitempty"`
	StartDate          string           `json:"start_date,omitempty"`
	EndDate            string           `json:"end_date,omitempty"`
	CustomCategories   []CustomCategory `json:"custom_categories,omitempty"`
	CustomInstructions string           `json:"custom_instructions,omitempty"`
	Messages           []Message        `json:"messages,omitempty"`
}

// CustomCategory 定义自定义类别
type CustomCategory map[string]any

// Message 定义消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MemoryHistory 定义内存历史
type MemoryHistory struct {
	ID         string    `json:"id"`
	MemoryID   string    `json:"memory_id"`
	Input      []Message `json:"input"`
	OldMemory  string    `json:"old_memory,omitempty"`
	NewMemory  string    `json:"new_memory,omitempty"`
	UserID     string    `json:"user_id"`
	Categories []string  `json:"categories"`
	Event      string    `json:"event"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Memory 定义内存
type Memory struct {
	ID         string         `json:"id"`
	Messages   []Message      `json:"messages,omitempty"`
	Event      string         `json:"event,omitempty"`
	Data       *MemoryData    `json:"data,omitempty"`
	Memory     string         `json:"memory,omitempty"`
	UserID     string         `json:"user_id,omitempty"`
	Hash       string         `json:"hash,omitempty"`
	Categories []string       `json:"categories,omitempty"`
	CreatedAt  time.Time      `json:"created_at,omitempty"`
	UpdatedAt  time.Time      `json:"updated_at,omitempty"`
	MemoryType string         `json:"memory_type,omitempty"`
	Score      float64        `json:"score,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Owner      string         `json:"owner,omitempty"`
	AgentID    string         `json:"agent_id,omitempty"`
	AppID      string         `json:"app_id,omitempty"`
	RunID      string         `json:"run_id,omitempty"`
}

// MemoryData 定义内存数据
type MemoryData struct {
	Memory string `json:"memory"`
}

// MemoryUpdateBody 定义内存更新请求体
type MemoryUpdateBody struct {
	MemoryID string `json:"memoryId"`
	Text     string `json:"text"`
}

// User 定义用户
type User struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	TotalMemories int       `json:"total_memories"`
	Owner         string    `json:"owner"`
	Type          string    `json:"type"`
}

// AllUsers 定义用户列表
type AllUsers struct {
	Count    int    `json:"count"`
	Results  []User `json:"results"`
	Next     any    `json:"next"`
	Previous any    `json:"previous"`
}

// ProjectResponse 定义项目响应
type ProjectResponse struct {
	CustomInstructions string   `json:"custom_instructions,omitempty"`
	CustomCategories   []string `json:"custom_categories,omitempty"`
}

// PromptUpdatePayload 定义提示更新请求体
type PromptUpdatePayload struct {
	CustomInstructions string           `json:"custom_instructions,omitempty"`
	CustomCategories   []CustomCategory `json:"custom_categories,omitempty"`
}

// WebhookEvent 定义 Webhook 事件类型
type WebhookEvent string

const (
	MemoryAdded   WebhookEvent = "memory_add"
	MemoryUpdated WebhookEvent = "memory_update"
	MemoryDeleted WebhookEvent = "memory_delete"
)

// Webhook 定义 Webhook
type Webhook struct {
	WebhookID  string         `json:"webhook_id,omitempty"`
	Name       string         `json:"name"`
	URL        string         `json:"url"`
	Project    string         `json:"project,omitempty"`
	CreatedAt  time.Time      `json:"created_at,omitempty"`
	UpdatedAt  time.Time      `json:"updated_at,omitempty"`
	IsActive   bool           `json:"is_active,omitempty"`
	EventTypes []WebhookEvent `json:"event_types,omitempty"`
}

// WebhookPayload 定义 Webhook 请求体
type WebhookPayload struct {
	EventTypes []WebhookEvent `json:"eventTypes"`
	ProjectID  string         `json:"projectId"`
	WebhookID  string         `json:"webhookId"`
	Name       string         `json:"name"`
	URL        string         `json:"url"`
}

// FeedbackPayload 定义反馈请求体
type FeedbackPayload struct {
	MemoryID       string   `json:"memory_id"`
	Feedback       Feedback `json:"feedback,omitempty"`
	FeedbackReason string   `json:"feedback_reason,omitempty"`
}

// SearchOptions 定义搜索选项
type SearchOptions struct {
	MemoryOptions
	Limit                   int      `json:"limit,omitempty"`
	EnableGraph             bool     `json:"enable_graph,omitempty"`
	Threshold               float64  `json:"threshold,omitempty"`
	TopK                    int      `json:"top_k,omitempty"`
	OnlyMetadataBasedSearch bool     `json:"only_metadata_based_search,omitempty"`
	KeywordSearch           bool     `json:"keyword_search,omitempty"`
	Fields                  []string `json:"fields,omitempty"`
	Categories              []string `json:"categories,omitempty"`
	Rerank                  bool     `json:"rerank,omitempty"`
}

// ProjectOptions 定义项目选项
type ProjectOptions struct {
	Fields []string `json:"fields,omitempty"`
}

// ToQuery 将结构体转换为 URL 查询字符串
func (o MemoryOptions) ToQuery() string {
	return structToQuery(o)
}

// ToQuery 将结构体转换为 URL 查询字符串
func (o SearchOptions) ToQuery() string {
	return structToQuery(o)
}

// ToQuery 将结构体转换为 URL 查询字符串
func (o ProjectOptions) ToQuery() string {
	return structToQuery(o)
}

// structToQuery 将结构体转换为 URL 查询字符串
func structToQuery(v interface{}) string {
	values := url.Values{}
	val := reflect.ValueOf(v)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		// 跳过零值
		if value.IsZero() {
			continue
		}

		// 获取 JSON 标签
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		// 处理字段值
		var strValue string
		switch value.Kind() {
		case reflect.String:
			strValue = value.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			strValue = strconv.FormatInt(value.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			strValue = strconv.FormatUint(value.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			strValue = strconv.FormatFloat(value.Float(), 'f', -1, 64)
		case reflect.Bool:
			strValue = strconv.FormatBool(value.Bool())
		case reflect.Slice:
			if value.Len() == 0 {
				continue
			}
			// 处理切片类型的字段
			var items []string
			for j := 0; j < value.Len(); j++ {
				item := value.Index(j)
				switch item.Kind() {
				case reflect.String:
					items = append(items, item.String())
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					items = append(items, strconv.FormatInt(item.Int(), 10))
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					items = append(items, strconv.FormatUint(item.Uint(), 10))
				case reflect.Float32, reflect.Float64:
					items = append(items, strconv.FormatFloat(item.Float(), 'f', -1, 64))
				case reflect.Bool:
					items = append(items, strconv.FormatBool(item.Bool()))
				}
			}
			strValue = strings.Join(items, ",")
		case reflect.Struct:
			if value.Type() == reflect.TypeOf(time.Time{}) {
				strValue = value.Interface().(time.Time).Format(time.RFC3339)
			} else {
				// 对于其他结构体类型，尝试转换为 JSON
				if jsonData, err := json.Marshal(value.Interface()); err == nil {
					strValue = string(jsonData)
				}
			}
		case reflect.Map:
			if value.Len() == 0 {
				continue
			}
			// 对于 map 类型，尝试转换为 JSON
			if jsonData, err := json.Marshal(value.Interface()); err == nil {
				strValue = string(jsonData)
			}
		default:
			continue
		}

		values.Add(tag, strValue)
	}

	return values.Encode()
}
