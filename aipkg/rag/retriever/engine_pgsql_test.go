package retriever

import (
	"context"
	"fmt"
	"testing"

	"github.com/jettjia/go-pkg/aipkg/schema"
	"github.com/jettjia/go-pkg/pkg/util"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getGorm() *gorm.DB {
	// 使用gorm配置数据库连接
	dsn := "host=192.168.64.126 user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 初始化表
	db.AutoMigrate(&schema.PgVector{})
	db.AutoMigrate(&schema.PgVectorDoc{})

	return db
}

// go test -v -run Test_pgsql_Save ./
func Test_pgsql_Save(t *testing.T) {
	db := getGorm()

	repo := NewPostgresRetrieveEngineRepository(db)

	// 创建包含embedding数据的params
	params1 := map[string]any{
		"embedding": map[string][]float32{
			"doc_1": {1, 2, 3},
		},
	}
	params2 := map[string]any{
		"embedding": map[string][]float32{
			"doc_1": {1, 2, 3},
		},
	}
	params3 := map[string]any{
		"embedding": map[string][]float32{
			"doc_1": {1, 2, 3},
		},
	}

	err := repo.Save(context.Background(), &schema.Document{
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

	// 从数据库中查询
	var pgVector schema.PgVector
	db.Where("doc_id = ?", "doc_1").First(&pgVector)
	if pgVector.DocID != "doc_1" {
		t.Fatal("doc_id not found")
	}
}

// go test -v -run Test_pgsql_SaveFullContent ./
func Test_pgsql_SaveFullContent(t *testing.T) {
	db := getGorm()

	repo := NewPostgresRetrieveEngineRepository(db)

	err := repo.SaveFullContent(context.Background(), &schema.Document{
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

// go test -v -run Test_pgsql_Retrieve ./
func Test_pgsql_Retrieve(t *testing.T) {
	db := getGorm()

	repo := NewPostgresRetrieveEngineRepository(db)

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
