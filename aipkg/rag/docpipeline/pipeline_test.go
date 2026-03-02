package docpipeline

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jettjia/go-pkg/aipkg/models/embedding"
	"github.com/jettjia/go-pkg/aipkg/pkg/ollama"
	"github.com/jettjia/go-pkg/aipkg/pkg/types"
	"github.com/jettjia/go-pkg/aipkg/rag/chunker"
	"github.com/panjf2000/ants/v2"
	"github.com/stretchr/testify/assert"
)

// go test -v -run TestPipeline_Process ./
func TestPipeline_Process(t *testing.T) {
	// 切片策略
	factory := chunker.GetStrategyFactory()
	strategy, flag := factory.GetStrategy(chunker.StrategyTypeRecursiveCharacter)
	if !flag {
		t.Errorf("获取策略 %s 失败: %s", chunker.StrategyTypeRecursiveCharacter, "non_existent")
	}

	// 创建embedding配置
	embeddingConfig := embedding.Config{
		APIKey:     os.Getenv("OPENAI_API_KEY"),
		BaseURL:    os.Getenv("OPENAI_BASE_URL"),
		ModelName:  "BAAI/bge-m3",
		Source:     types.ModelSourceRemote,
		ModelID:    "BAAI/bge-m3",
		Dimensions: 1024,
	}

	// 创建goroutine池用于embedding
	pool, err := ants.NewPool(10) // 创建一个大小为10的goroutine池
	if err != nil {
		t.Fatalf("Failed to create goroutine pool: %v", err)
	}
	defer pool.Release()

	// 创建BatchEmbedder实例
	batchEmbedder := embedding.NewBatchEmbedder(pool)

	// 创建OllamaService实例
	ollamaService, err := ollama.GetOllamaService()
	if err != nil {
		t.Fatalf("Failed to create Ollama service: %v", err)
	}

	// 创建embedder实例
	embedder, err := embedding.NewEmbedder(embeddingConfig, batchEmbedder, ollamaService)
	if err != nil {
		t.Fatalf("Failed to create embedder: %v", err)
	}

	// 创建pipeline处理器
	processor, err := NewDefaultDocumentProcessor(embedder, strategy, 4)
	if err != nil {
		t.Fatalf("Failed to create document processor: %v", err)
	}
	defer processor.Close()

	// 准备测试文本
	testText := `# **Kubernetes 应用场景全解析**\n\n在当今数字化时代，Kubernetes（K8s）作为容器编排领域的佼佼者，为各类应用的高效管理与部署提供了卓越的解决方案，在众多场景中发挥着关键作用。\n\n## **微服务架构管理**\n\n微服务架构将单一应用拆分为多个小型、独立的服务，各服务可独立开发、部署与扩展，极大提升了应用的灵活性与可维护性。而 K8s 堪称管理微服务架构的利器。它能够将不同微服务分别打包成容器，通过 “服务（Service）” 这一抽象层，实现微服务间的高效发现与通信。同时，借助 K8s 的负载均衡机制，可将请求均匀分发至各个微服务实例，保障系统的高可用性与性能。例如，一个电商平台中，商品展示、订单处理、用户管理等微服务均可部署在 K8s 集群中，各自独立运行，彼此协同工作，轻松应对业务的快速发展与变化。\n\n## **大规模应用部署与弹性扩展**\n\n对于大型网站或云计算应用而言，每日要处理海量用户请求。K8s 的自动伸缩功能（Horizontal Pod Autoscaler，HPA）在此时大显身手，它能依据设定的指标（如 CPU 利用率、内存使用量、请求数等），自动增加或减少容器实例数量。在电商大促、直播活动等流量高峰时段，K8s 可快速扩容应用实例，确保系统稳定运行；而在流量低谷时，则自动缩容，降低资源消耗与成本。以在线教育平台为例，在上课高峰期，K8s 自动增加视频播放、互动答题等服务的容器数量；课程结束后，又及时减少实例，实现资源的精准调配。\n\n## **自动化部署与持续集成 / 持续交付（CI/CD）**\n\n在软件开发流程里，CI/CD 是实现高效交付的关键环节。K8s 与 CI/CD 工具的深度集成，可实现应用的自动化部署。开发人员将代码提交至代码仓库后，CI/CD 流水线自动构建容器镜像，并推送至镜像仓库。随后，K8s 依据预先定义的配置，从镜像仓库拉取镜像，完成应用的部署。并且，K8s 支持滚动更新，在更新应用时，以逐步替换的方式更新容器，保障服务不间断；若更新过程出现问题，还能快速回滚到上一稳定版本，极大提升了软件交付的效率与可靠性。众多互联网企业通过这种方式，实现了每天多次的软件更新与发布。\n\n## **人工智能与机器学习工作负载管理**\n\n人工智能与机器学习领域的模型训练与推理任务，通常需要大量计算资源，且对资源的动态调配需求较高。K8s 能够将模型训练任务容器化，实现任务的高效调度与资源分配。借助 Kubeflow 等基于 K8s 的机器学习框架扩展，数据科学家可方便地管理模型训练、超参数调优、模型部署等全生命周期流程。例如，在图像识别项目中，利用 K8s 集群的强大算力，并行运行多个模型训练任务，大幅缩短训练时间，加速模型迭代。\n\n## **混合云和多云环境管理**\n\n随着企业数字化转型的不断深化，业务场景的复杂性与数据规模的激增，推动混合云（同时使用公有云与私有云资源）和多云（使用多个不同公有云服务）部署模式成为企业 IT 架构的主流选择。然而，这种架构也带来了新的挑战 —— 不同云平台的技术标准不统一、资源调度分散、运维成本攀升等问题，让企业在享受灵活部署优势的同时，面临着管理效率低下的困境。​\n\n在此背景下，Kubernetes（简称 K8s）凭借其跨平台兼容性与云厂商无关性的核心特性，成为破解混合云和多云管理难题的理想方案。企业可基于 AWS、Azure、阿里云等不同公有云平台，以及自建的私有云基础设施，分别构建独立的 K8s 集群，再通过 K8s 联邦（Kubernetes Federation）等工具实现多集群的统一管控。这种架构不仅能打破云平台间的技术壁垒，实现应用部署流程、资源监控指标、运维管理策略的标准化，还能让企业根据业务需求灵活分配资源，充分发挥各云平台的独特优势。​\n\n例如，对于金融、医疗等对数据安全性要求极高的行业，可将用户敏感数据存储、核心交易处理等任务部署在私有云 K8s 集群中，依托私有云的专属网络与访问控制机制保障数据安全；而对于电商大促、在线直播等需要应对高并发流量的场景，则可将前端应用、负载均衡服务部署在公有云 K8s 集群，借助公有云弹性扩容能力快速应对流量波动，避免资源闲置与业务中断风险。​\n\n此外，K8s 的自动扩缩容、滚动更新等功能，还能进一步提升混合云和多云环境的运维效率。当某一云平台资源紧张时，K8s 可自动将部分任务调度至资源充足的集群；在应用更新时，滚动更新功能能确保业务不中断，极大降低了运维成本与风险。同时，K8s 的开放性也让企业无需依赖单一云厂商，有效避免 “厂商锁定” 问题，未来可根据成本、性能等因素灵活切换或新增云服务，进一步提升资源使用的灵活性与成本效益，为企业数字化转型提供稳定、高效的 IT 架构支撑。`

	// 配置处理选项
	processOptions := ProcessOptions{
		EnableConcurrency: false,
	}

	// 执行处理流程
	documents, err := processor.Process(context.Background(), testText, processOptions)
	if err != nil {
		t.Fatalf("Failed to process document: %v", err)
	}

	// 验证结果
	assert.NotNil(t, documents, "Documents should not be nil")
	assert.Greater(t, len(documents), 0, "Should generate at least one document")

	// 打印结果信息
	fmt.Printf("Generated %d documents\n", len(documents))
	for i, doc := range documents {
		fmt.Printf("Document %d:\n", i+1)
		fmt.Printf("  Title: %s\n", doc.Title)
		fmt.Printf("  Depth: %d\n", doc.Depth)
		fmt.Printf("  Content length: %v\n", doc.Content)
		fmt.Printf("  ChunkId: %s\n", doc.ChunkId)
		// fmt.Printf("  Has embedding: %v\n", doc.Embedding)
		fmt.Printf("  Page: %v\n", doc.Page)
		fmt.Printf("  Has embedding: %v\n", len(doc.Embedding) > 0)
	}
}

// go test -v -run TestPipeline_BatchProcess ./
func TestPipeline_BatchProcess(t *testing.T) {
	// 切片策略
	factory := chunker.GetStrategyFactory()
	strategy, flag := factory.GetStrategy(chunker.StrategyTypeRecursiveCharacter)
	if !flag {
		t.Errorf("获取策略 %s 失败: %s", chunker.StrategyTypeRecursiveCharacter, "non_existent")
	}

	// 创建embedding配置
	embeddingConfig := embedding.Config{
		APIKey:     os.Getenv("OPENAI_API_KEY"),
		BaseURL:    os.Getenv("OPENAI_BASE_URL"),
		ModelName:  "BAAI/bge-m3",
		Source:     types.ModelSourceRemote,
		ModelID:    "BAAI/bge-m3",
		Dimensions: 1024,
	}

	// 创建goroutine池用于embedding
	pool, err := ants.NewPool(10)
	if err != nil {
		t.Fatalf("Failed to create goroutine pool: %v", err)
	}
	defer pool.Release()

	// 创建BatchEmbedder实例
	batchEmbedder := embedding.NewBatchEmbedder(pool)

	// 创建OllamaService实例
	ollamaService, err := ollama.GetOllamaService()
	if err != nil {
		t.Fatalf("Failed to create Ollama service: %v", err)
	}

	// 创建embedder实例
	embedder, err := embedding.NewEmbedder(embeddingConfig, batchEmbedder, ollamaService)
	if err != nil {
		t.Fatalf("Failed to create embedder: %v", err)
	}

	// 创建pipeline处理器
	processor, err := NewDefaultDocumentProcessor(embedder, strategy, 4)
	if err != nil {
		t.Fatalf("Failed to create document processor: %v", err)
	}
	defer processor.Close()

	// 准备多个测试文本
	testTexts := []string{
		`# 文档1,这是第一个测试文档。## 内容1,这是文档1的内容。`,
		`# 文档2,这是第二个测试文档。## 内容2,这是文档2的内容。`,
	}

	// 配置处理选项（启用并发）
	processOptions := ProcessOptions{
		EnableConcurrency: true,
		MaxConcurrency:    2,
	}

	// 执行批量处理流程
	results, err := processor.BatchProcess(context.Background(), testTexts, processOptions)
	if err != nil {
		t.Fatalf("Failed to batch process documents: %v", err)
	}

	// 验证结果
	assert.NotNil(t, results, "Results should not be nil")
	assert.Equal(t, len(testTexts), len(results), "Results length should match texts length")

	// 打印结果信息
	for i, docs := range results {
		fmt.Printf("Document set %d generated %d chunks\n", i+1, len(docs))
		for j, doc := range docs {
			fmt.Printf("  Chunk %d: Title='%s', Content length=%d, Has embedding=%v\n",
				j+1, doc.Title, len(doc.Content), doc.Embedding != nil)
		}
	}
}

// go test -v -run TestPipeline_SliceOnly ./
func TestPipeline_SliceOnly(t *testing.T) {
	// 切片策略
	factory := chunker.GetStrategyFactory()
	strategy, flag := factory.GetStrategy(chunker.StrategyTypeRecursiveCharacter)
	if !flag {
		t.Errorf("获取策略 %s 失败: %s", chunker.StrategyTypeRecursiveCharacter, "non_existent")
	}

	// 创建pipeline处理器
	processor, err := NewDefaultDocumentProcessor(nil, strategy, 4)
	if err != nil {
		t.Fatalf("Failed to create document processor: %v", err)
	}
	defer processor.Close()

	// 准备测试文本
	testText := `# 文档1,这是第一个测试文档。## 内容1,这是文档1的内容。`

	// 配置处理选项
	processOptions := ProcessOptions{
		EnableConcurrency: false,
	}

	// 执行切片流程
	documents, err := processor.SliceOnly(context.Background(), testText, processOptions)
	if err != nil {
		t.Fatalf("Failed to slice document: %v", err)
	}

	// 验证结果
	assert.NotNil(t, documents, "Documents should not be nil")
	assert.Greater(t, len(documents), 0, "Should generate at least one document")

	// 打印结果信息
	fmt.Printf("Generated %d documents\n", len(documents))
	for i, doc := range documents {
		fmt.Printf("Document %d:\n", i+1)
		fmt.Printf("  Title: %s\n", doc.Title)
		fmt.Printf("  Depth: %d\n", doc.Depth)
		fmt.Printf("  Content length: %v\n", doc.Content)
		fmt.Printf("  ChunkId: %s\n", doc.ChunkId)
		fmt.Printf("  Has embedding: %v\n", doc.Embedding)
		// fmt.Printf("  Has embedding: %v\n", len(doc.Embedding) > 0)
	}
}
