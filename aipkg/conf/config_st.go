package conf

type Config struct {
	Ai AiConf
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
