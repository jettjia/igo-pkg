package conf

import (
	"encoding/json"
	"errors"
)

type Config struct {
	InitServer InitConf
	Server     ServerConf
	Gserver    GServerConf
	DB         DBConf
	DBManager  DBManagerConf `yaml:"db_manager"`
	Log        LogConf
	Mq         MqConf
	Otel       OtelConf
	Cache      CacheConf
	Search     SearchConf
	Third      ThirdConf
	Ai         AiConf
}

type InitConf struct {
	Initdsn string `yaml:"initdsn"`
}

type ServerConf struct {
	Lang        string `yaml:"lang"`
	PublicPort  int    `yaml:"public_port"`
	PrivatePort int    `yaml:"private_port"`
	ServerName  string `yaml:"server_name"`
	Mode        string `yaml:"mode"`
	Dev         bool   `yaml:"dev"`
	EnableEvent bool   `yaml:"enable_event"`
	EnableJob   bool   `yaml:"enable_job"`
	EnableGrpc  bool   `yaml:"enable_grpc"`
	EnableMcp   bool   `yaml:"enable_mcp"`

	// mcp
	McpPublicPort int `yaml:"mcp_public_port"`

	// other
	PublicMetricPort  int    `yaml:"public_metric_port"`
	PrivateMetricPort int    `yaml:"private_metric_port"`
	XosPort           int    `yaml:"xos_port"`
	S3Region          string `yaml:"s3Region"`
	Checks3           bool   `yaml:"check_s3"`
}

type GServerConf struct {
	Host            string `yaml:"host"`
	PublicPort      int    `yaml:"public_port"`
	MaxMsgSize      int    `yaml:"max_msg_size"`
	ClientGoodsHost string `yaml:"client_goods_host"`
	ClientGoodsPort int    `yaml:"client_goods_port"`
}

type LogConf struct {
	LogFileDir string `yaml:"log_file_dir"` // Log Directory
	AppName    string `yaml:"app_name"`     // Log Name
	MaxSize    int    `yaml:"max_size"`     // At What File Size Should Log Rotation Start
	MaxBackups int    `yaml:"max_backups"`  // Number of Retained Files
	MaxAge     int    `yaml:"max_age"`      // Maximum File Retention Time
	LogLevel   string `yaml:"log_level"`    // Log Levels
	LogOut     int    `yaml:"log_out"`      // Log Output Methods
}

// DBConf database config like pg/mysql...
type DBConf struct {
	DbType          string `yaml:"db_type"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	DbHost          string `yaml:"db_host"`
	DbPort          int    `yaml:"db_port"`
	DbName          string `yaml:"db_name"`
	Charset         string `yaml:"charset"`
	MaxIdleConn     int    `yaml:"max_idle_conn"`
	MaxOpenConn     int    `yaml:"max_open_conn"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
	LogMode         int    `yaml:"log_mode"`
	SlowThreshold   int    `yaml:"slow_threshold"`
}

// DBManager db manager
type DBManagerConf struct {
	DataSources []DataSourceCfg `yaml:"data_sources"`
}

type DataSourceCfg struct {
	DbType    string   `yaml:"db_type"`    // Database Types
	Name      string   `yaml:"name"`       // Data Source Name
	MasterDSN string   `yaml:"master_dsn"` // Primary Node DSN
	SlaveDSNs []string `yaml:"slave_dsns"` // Secondary Node DSN

	// It can also be left unset, as there will be default values.
	MaxIdleConn   int // Max Idle Connections
	MaxOpenConn   int // Max Connections
	MaxLifetime   int // Maximum Lifetime (s)
	LogMode       int // Log Level(gorm; 1: Silent, 2:Error,3:Warn,4:Info)
	SlowThreshold int // Slow SQL Judgment Time (s)
}

// CacheConf redis
type CacheConf struct {
	CacheType string `yaml:"cache_type"` // pgsql/redis
	Addr      string `yaml:"addr"`
	Password  string `yaml:"password"`
	// Redis Config
	RedisType  string `yaml:"redis_type"` // alone, sentinel,cluster
	MasterName string `yaml:"master_name"`
	PoolSize   int    `yaml:"pool_size"`
}

// MqConf mq
type MqConf struct {
	MqType          string `yaml:"mq_type"` // pgsql/memory/redis/redis-stream
	MqProducerHost  string `yaml:"mq_producer_host"`
	MqProducerPort  int    `yaml:"mq_producer_port"`
	MqSubscribeHost string `yaml:"mq_subscribe_host"`
	MqSubscribePort int    `yaml:"mq_subscribe_port"`
}

// SearchConf search
type SearchConf struct {
	SearchType string `yaml:"search_type"` // es, meilisearch, manticoresearch
	Addr       string `yaml:"addr"`        // http://127.0.0.1:9200, http://127.0.0.1:7700, http://127.0.0.1:9308
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
}

// OtelConf otel
type OtelConf struct {
	Enable         bool   `yaml:"enable"`
	ExportEndpoint string `yaml:"export_endpoint"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
}

// Aiconf
type AiConf struct {
	ExtractEntitiesPrompt      string `yaml:"extract_entities_prompt" json:"extract_entities_prompt"`           // 实体提取
	ExtractRelationshipsPrompt string `yaml:"extract_relationships_prompt" json:"extract_relationships_prompt"` // 关系提取
	KeywordsExtractionPrompt   string `yaml:"keywords_extraction_prompt" json:"keywords_extraction_prompt"`     // 提问关键词提取
	RewritePromptSystem        string `yaml:"rewrite_prompt_system" json:"rewrite_prompt_system"`               // 提问重写系统提示词
	RewritePromptUser          string `yaml:"rewrite_prompt_user" json:"rewrite_prompt_user"`                   // 提问重写用户提示词
	NL2SQLPromptSystem         string `yaml:"nl2sql_prompt_system" json:"nl2sql_prompt_system"`                 //  nl2sql 系统提示词
	NL2SQLPromptUser           string `yaml:"nl2sql_prompt_user" json:"nl2sql_prompt_user"`                     //  nl2sql 用户提示词
}

// ThirdConf third-party services
type ThirdConf struct {
	Extra map[string]interface{}
}

// UnmarshalYAML
func (t *ThirdConf) UnmarshalYAML(unmarshal func(interface{}) error) error {
	t.Extra = make(map[string]interface{})
	var m map[string]interface{}
	if err := unmarshal(&m); err != nil {
		return err
	}
	for k, v := range m {
		t.Extra[k] = v
	}
	return nil
}

// UnmarshalJSON
func (t *ThirdConf) UnmarshalJSON(data []byte) error {
	t.Extra = make(map[string]interface{})
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	for k, v := range m {
		t.Extra[k] = v
	}
	return nil
}

// UnmarshalTOML
func (t *ThirdConf) UnmarshalTOML(data interface{}) error {
	t.Extra = make(map[string]interface{})
	m, ok := data.(map[string]interface{})
	if !ok {
		return errors.New("toml data is not a map")
	}
	for k, v := range m {
		t.Extra[k] = v
	}
	return nil
}
