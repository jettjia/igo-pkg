package docpipeline

import (
	"context"
	"errors"
	"sync"

	"github.com/jettjia/igo-pkg/aipkg/models/embedding"
	"github.com/jettjia/igo-pkg/aipkg/rag/chunker"
	"github.com/jettjia/igo-pkg/aipkg/schema"
	"github.com/panjf2000/ants/v2"
)

// DefaultDocumentProcessor 默认文档处理器
type DefaultDocumentProcessor struct {
	strategy   chunker.ChunkingStrategy
	embedder   embedding.Embedder
	workerPool *ants.Pool
	mutex      sync.RWMutex
}

// NewDefaultDocumentProcessor 创建新的默认文档处理器
func NewDefaultDocumentProcessor(embedder embedding.Embedder, chunkingStrategy chunker.ChunkingStrategy, maxWorkers int) (*DefaultDocumentProcessor, error) {
	if maxWorkers <= 0 {
		maxWorkers = 4 // 设置默认工作线程数
	}

	// 创建工作池
	wp, err := ants.NewPool(maxWorkers)
	if err != nil {
		return nil, err
	}

	// 返回新的文档处理器实例
	return &DefaultDocumentProcessor{
		strategy:   chunkingStrategy,
		embedder:   embedder,
		workerPool: wp,
	}, nil
}

// ErrPartialVectorizationFailed 表示部分切片向量化失败的错误
var ErrPartialVectorizationFailed = errors.New("部分切片向量化失败")

// Process 实现文档处理接口，优化后支持部分切片向量化失败的处理
func (p *DefaultDocumentProcessor) Process(ctx context.Context, text string, options ProcessOptions) ([]*schema.Document, error) {
	// 1. 切片处理
	docs, err := p.strategy.Chunk(ctx, text)
	if err != nil {
		return nil, err
	}

	// 2. 过滤空文档并提取需要向量化的文本
	var texts []string
	var nonEmptyDocs []*schema.Document
	var emptyDocIndices []int

	for i, doc := range docs {
		if doc.Content != "" {
			texts = append(texts, doc.Content)
			nonEmptyDocs = append(nonEmptyDocs, doc)
		} else {
			// 记录空文档的索引
			emptyDocIndices = append(emptyDocIndices, i)
		}
	}

	// 3. 处理空文档
	for _, idx := range emptyDocIndices {
		if docs[idx].MetaData == nil {
			docs[idx].MetaData = make(map[string]any)
		}
		docs[idx].MetaData["empty_content"] = true
		docs[idx].MetaData["vectorization_failed"] = true
		docs[idx].MetaData["vectorization_error"] = "empty content cannot be vectorized"
	}

	// 4. 如果没有非空文档，直接返回
	if len(nonEmptyDocs) == 0 {
		return docs, nil
	}

	// 5. 为每个非空文档创建或初始化Metadata
	for i := range nonEmptyDocs {
		if nonEmptyDocs[i].MetaData == nil {
			nonEmptyDocs[i].MetaData = make(map[string]any)
		}
	}

	// 6. 批量向量化
	vectors, err := p.embedder.BatchEmbed(ctx, texts)

	// 7. 如果批量向量化失败，尝试单独向量化每个非空切片
	var hasFailed bool
	if err != nil {
		// 初始化vectors数组
		vectors = make([][]float32, len(nonEmptyDocs))
		hasFailed = true

		// 单独向量化每个非空切片
		for i := range nonEmptyDocs {
			// 使用文档的Content属性而不是texts数组，保持一致性
			vector, err := p.embedder.Embed(ctx, nonEmptyDocs[i].Content)
			if err != nil {
				// 记录向量化失败的信息
				nonEmptyDocs[i].MetaData["vectorization_failed"] = true
				nonEmptyDocs[i].MetaData["vectorization_error"] = err.Error()
				// 保留空向量，便于业务方识别
				vectors[i] = nil
			} else {
				vectors[i] = vector
			}
		}
	}

	// 8. 将向量赋值给非空文档
	for i, doc := range nonEmptyDocs {
		if i < len(vectors) && vectors[i] != nil {
			doc.Embedding = vectors[i]
			doc.MetaData["has_embedding"] = true
		}
	}

	// 7. 如果有部分切片向量化失败，返回自定义错误
	if hasFailed {
		return docs, ErrPartialVectorizationFailed
	}

	return docs, nil
}

// BatchProcess 批量处理文档
func (p *DefaultDocumentProcessor) BatchProcess(ctx context.Context, texts []string, options ProcessOptions) ([][]*schema.Document, error) {
	results := make([][]*schema.Document, len(texts))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	// 非并发模式
	if !options.EnableConcurrency {
		for i, text := range texts {
			docs, err := p.Process(ctx, text, options)
			if err != nil {
				return nil, err
			}
			results[i] = docs
		}
		return results, nil
	}

	// 并发模式
	// 如果设置了MaxConcurrency且大于0，则使用该值限制并发数
	var semaphore chan struct{}
	if options.MaxConcurrency > 0 {
		semaphore = make(chan struct{}, options.MaxConcurrency)
	}

	for i, text := range texts {
		if firstErr != nil {
			break
		}
		wg.Add(1)
		idx := i
		text := text

		// 如果设置了semaphore，则获取信号量
		if semaphore != nil {
			semaphore <- struct{}{}
		}

		err := p.workerPool.Submit(func() {
			// 如果使用了semaphore，完成后释放信号量
			if semaphore != nil {
				defer func() {
					<-semaphore
				}()
			}
			defer wg.Done()
			if firstErr != nil {
				return
			}
			docs, err := p.Process(ctx, text, options)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			results[idx] = docs
		})
		if err != nil {
			// 如果提交任务失败，需要释放已获取的信号量
			if semaphore != nil {
				<-semaphore
			}
			return nil, err
		}
	}

	wg.Wait()
	return results, firstErr
}

// Close 释放资源
func (p *DefaultDocumentProcessor) Close() {
	p.workerPool.Release()
}

// SliceOnly 只切片，使用recursive_character策略进行分块
func (p *DefaultDocumentProcessor) SliceOnly(ctx context.Context, text string, options ProcessOptions) ([]*schema.Document, error) {
	// 参数验证
	if p.strategy == nil {
		return nil, errors.New("chunking strategy is nil")
	}

	// 1. 切片处理
	docs, err := p.strategy.Chunk(ctx, text)
	if err != nil {
		return nil, err
	}

	// 2. 处理空文档，为其添加元数据标识
	for _, doc := range docs {
		if doc.Content == "" {
			if doc.MetaData == nil {
				doc.MetaData = make(map[string]any)
			}
			doc.MetaData["empty_content"] = true
		} else if doc.MetaData == nil {
			// 为非空文档初始化元数据
			doc.MetaData = make(map[string]any)
		}
	}

	return docs, nil
}
