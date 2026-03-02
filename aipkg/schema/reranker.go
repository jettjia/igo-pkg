package schema

type RerankRequest struct {
	Query string
	Data  [][]*RerankData
	TopN  *int64
}

type RerankResponse struct {
	SortedData []*RerankData
	TokenUsage *int64
}

type RerankData struct {
	Document *Document
	Score    float64 // 原始分数
	RRFScore float64 // RRF计算后的分数
}
