package retriever

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jettjia/igo-pkg/aipkg/schema"
	"github.com/jettjia/igo-pkg/pkg/util"
)

// 初始化OpenSearch客户端
func getOpenSearchClient() (RetrieveEngineRepository, error) {

	client, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{os.Getenv("ELASTICSEARCH_ADDR")},
		Username:  os.Getenv("ELASTICSEARCH_USERNAME"),
		Password:  os.Getenv("ELASTICSEARCH_PASSWORD"),
	})
	if err != nil {
		return nil, err
	}

	// 创建OpenSearch检索引擎仓库
	repo := NewElasticsearchEngineRepository(client)

	return repo, nil
}

// go test -v -run Test_opensearch_Save ./
func Test_opensearch_Save(t *testing.T) {
	// 获取OpenSearch客户端
	repo, err := getOpenSearchClient()
	if err != nil {
		t.Fatal(err)
	}

	// 创建包含embedding数据的params
	params1 := map[string]any{
		"embedding": map[string][]float32{
			"doc_1": {1, 2, 3},
		},
	}
	params2 := map[string]any{
		"embedding": map[string][]float32{
			"doc_1": {4, 5, 6},
		},
	}
	params3 := map[string]any{
		"embedding": map[string][]float32{
			"doc_1": {7, 8, 9},
		},
	}

	// 保存文档
	err = repo.Save(context.Background(), &schema.Document{
		KnowledgeID: "knowledge_1",
		ChunkId:     "chunk_1",
		DocID:       "doc_1",
		Content:     "[测试文档1][切片1]这是一个测试文档，包含测试关键词，用于测试全文搜索",
	}, params1)
	if err != nil {
		t.Fatal(err)
	}

	err = repo.Save(context.Background(), &schema.Document{
		KnowledgeID: "knowledge_2",
		ChunkId:     "chunk_2",
		DocID:       "doc_1",
		Content:     "[测试文档1][切片2]这是另一个文档，不包含任何关键词",
	}, params2)
	if err != nil {
		t.Fatal(err)
	}

	err = repo.Save(context.Background(), &schema.Document{
		KnowledgeID: "knowledge_3",
		ChunkId:     "chunk_3",
		DocID:       "doc_1",
		Content:     "[测试文档1][切片3]测试测试测试，重要的事情说三遍，这是一个高度相关的测试文档",
	}, params3)
	if err != nil {
		t.Fatal(err)
	}

	// 验证保存结果
	err = repo.Save(context.Background(), &schema.Document{
		KnowledgeID: "knowledge_1",
		ChunkId:     "chunk_1",
		DocID:       "doc_1",
		Content:     "[测试文档1]这是一个测试文档，包含测试关键词，用于测试全文搜索，我是整篇文档的内容",
	}, params1)
	if err != nil {
		t.Fatal(err)
	}
}

// go test -v -run Test_opensearch_SaveFullContent ./
func Test_opensearch_SaveFullContent(t *testing.T) {
	// 获取OpenSearch客户端
	repo, err := getOpenSearchClient()
	if err != nil {
		t.Fatal(err)
	}

	// 保存完整文档
	err = repo.SaveFullContent(context.Background(), &schema.Document{
		KnowledgeID: "knowledge_1",
		ChunkId:     "chunk_1",
		DocID:       "doc_1",
		Content:     "[测试文档1]这是一个测试文档，包含测试关键词，用于测试全文搜索，我是整篇文档的内容",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = repo.SaveFullContent(context.Background(), &schema.Document{
		KnowledgeID: "knowledge_1",
		ChunkId:     "chunk_1",
		DocID:       "doc_2",
		Content:     "[测试文档2]这是一个测试文档，包含测试关键词，用于测试全文搜索，我是整篇文档的内容，这是一个测试文档，包含测试关键词，用于测试全文搜索，我是整篇文档的内容",
	})
	if err != nil {
		t.Fatal(err)
	}
}

// go test -v -run Test_opensearch_Retrieve ./
func Test_opensearch_Retrieve(t *testing.T) {
	// 获取OpenSearch客户端
	repo, err := getOpenSearchClient()
	if err != nil {
		t.Fatal(err)
	}

	// 执行检索
	results, err := repo.Retrieve(context.Background(), schema.RetrieveParams{
		Query:         "测试",
		Embedding:     []float32{1, 2, 3},
		TopK:          4,
		Threshold:     0.05,
		RetrieverType: schema.HybridRetrieverType,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("retrieve result not found")
	}

	fmt.Println(util.PrintJson(results))
}
