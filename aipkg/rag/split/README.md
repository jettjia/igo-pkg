# Splitter 使用指南

文本分块器，支持 4 种分块策略，返回 `[]*Chunk` 切片。

---

## 返回值格式

### Chunk 结构

```json
{
  "doc_name": "example.md",
  "doc_md5": "a1b2c3d4e5f6...",
  "slice_md5": "f6e5d4c3b2a1...",
  "id": "chunk-00001",
  "pages": [1, 2],
  "segment_id": 1,
  "superior_id": "",
  "slice_content": {
    "title": "标题",
    "text": "文本内容...",
    "table": "",
    "picture": ""
  }
}
```

| 字段                  | 类型   | 说明                      |
| --------------------- | ------ | ------------------------- |
| doc_name              | string | 传入的文件名称            |
| doc_md5               | string | 原文的 MD5 值             |
| slice_md5             | string | 当前切片的 MD5 值         |
| id                    | string | 切片唯一 ID               |
| pages                 | []int  | 切片所在页码              |
| segment_id            | int    | 切片序号（按阅读顺序）    |
| superior_id           | string | 父切片 ID（用于层级结构） |
| slice_content.title   | string | 标题内容                  |
| slice_content.text    | string | 切片文本内容              |
| slice_content.table   | string | 表格内容（Markdown 格式） |
| slice_content.picture | string | 图片路径                  |

---

## 通用配置 (StrategyBase)

所有策略都继承以下配置：

| 参数                | 类型    | 说明                     | 默认值 |
| ------------------- | ------- | ------------------------ | ------ |
| ChunkSize           | int     | 单个块最大字符数         | 500    |
| OverlapRatio        | float64 | 相邻块重叠比例 (0.1~0.2) | 0.1    |
| RemoveURLAndEmail   | bool    | 是否移除 URL 和邮箱      | false  |
| NormalizeWhitespace | bool    | 是否规范化空白字符       | false  |
| TrimSpace           | bool    | 是否去除首尾空白         | false  |

---

## 1. FixedSizeStrategy (固定大小分块)

按固定字符数将文本分割成块。

### 创建方式

```go
s := split.NewFixedSizeStrategy()
s.ChunkSize = 400
s.ChunkOverlap = 50
s.OverlapRatio = 0.1
s.NormalizeWhitespace = true
s.TrimSpace = true
```

### 特有参数

| 参数         | 类型 | 说明                                    | 默认值 |
| ------------ | ---- | --------------------------------------- | ------ |
| ChunkOverlap | int  | 相邻块重叠字符数（优先于 OverlapRatio） | 0      |

### 调用示例

```go
ctx := context.Background()
text := "要分块的文本内容..."
fileName := "example.md"

s := split.NewFixedSizeStrategy()
s.ChunkSize = 400
s.ChunkOverlap = 50

chunks, err := s.Split(ctx, text, fileName)
if err != nil {
    panic(err)
}

for _, chunk := range chunks {
    fmt.Println(chunk.SliceContent.Text)
}
```

---

## 2. RecursiveCharacterStrategy (递归字符分块)

按分隔符层级递归分割文本，保持语义完整性。

### 创建方式

```go
s := split.NewRecursiveCharacterStrategy()
s.ChunkSize = 400
s.OverlapRatio = 0.15
s.RemoveURLAndEmail = false
s.NormalizeWhitespace = true
s.TrimSpace = true
```

### 调用示例

```go
ctx := context.Background()
text := "要分块的文本内容..."
fileName := "example.md"

s := split.NewRecursiveCharacterStrategy()
s.ChunkSize = 400
s.OverlapRatio = 0.15

chunks, err := s.Split(ctx, text, fileName)
if err != nil {
    panic(err)
}

for _, chunk := range chunks {
    fmt.Println(chunk.SliceContent.Text)
}
```

---

## 3. SemanticStrategy (语义分块)

基于 TF-IDF 余弦相似度进行语义分块。

### 创建方式

```go
s := split.NewSemanticStrategy()
s.ChunkSize = 400
s.Mode = split.SemanticModeDoubleMerging
s.InitialThreshold = 0.2
s.AppendingThreshold = 0.3
s.MergingThreshold = 0.3
s.MaxChunkSize = 400
s.MinChunkSize = 120
s.MergingRange = 2
s.MergingSeparator = ""
```

### 特有参数

| 参数                 | 类型         | 说明                                                        | 默认值 |
| -------------------- | ------------ | ----------------------------------------------------------- | ------ |
| Mode                 | SemanticMode | 语义模式：`greedy` / `window_breakpoint` / `double_merging` | greedy |
| Threshold            | float64      | 相似度阈值 (0~1)                                            | 0.5    |
| BufferSize           | int          | 窗口模式缓冲区大小                                          | 1      |
| BreakpointPercentile | int          | 断点百分位数 (50~99)                                        | 95     |
| MinChunkSize         | int          | 最小块大小                                                  | 0      |
| InitialThreshold     | float64      | 初始分块相似度阈值                                          | 0.2    |
| AppendingThreshold   | float64      | 追加句子相似度阈值                                          | 0.3    |
| MergingThreshold     | float64      | 合并块相似度阈值                                            | 0.3    |
| MaxChunkSize         | int          | 最大块大小                                                  | 0      |
| MergingRange         | int          | 合并范围 (1~2)                                              | 1      |
| MergingSeparator     | string       | 合并分隔符                                                  | ""     |

### SemanticMode 选项

| 模式                | 说明                               |
| ------------------- | ---------------------------------- |
| `greedy`            | 贪心模式，基于相似度阈值合并句子   |
| `window_breakpoint` | 窗口断点模式，计算滑动窗口相似度   |
| `double_merging`    | 双轮合并模式，先初始分块再合并优化 |

### 调用示例

```go
ctx := context.Background()
text := "要分块的文本内容..."
fileName := "example.md"

s := split.NewSemanticStrategy()
s.ChunkSize = 400
s.Mode = split.SemanticModeDoubleMerging
s.InitialThreshold = 0.2
s.AppendingThreshold = 0.3
s.MergingThreshold = 0.3
s.MaxChunkSize = 400
s.MinChunkSize = 120
s.MergingRange = 2

chunks, err := s.Split(ctx, text, fileName)
if err != nil {
    panic(err)
}

for _, chunk := range chunks {
    fmt.Println(chunk.SliceContent.Text)
}
```

---

## 4. DocumentStructureStrategy (文档结构分块)

根据 Markdown/HTML 标题层级结构进行分块。

### 创建方式

```go
s := split.NewDocumentStructureStrategy()
s.ChunkSize = 400
s.OverlapRatio = 0.1
s.MaxDepth = 3
s.SemanticThreshold = 0.5
s.SkipEmptyHeadings = true
```

### 特有参数

| 参数              | 类型    | 说明                 | 默认值 |
| ----------------- | ------- | -------------------- | ------ |
| SemanticThreshold | float64 | 语义相似度阈值 (0~1) | 0.5    |
| MaxDepth          | int     | 最大标题深度 (1~6)   | 3      |
| SkipEmptyHeadings | bool    | 是否跳过空标题       | true   |

### 调用示例

```go
ctx := context.Background()
text := "要分块的文本内容..."
fileName := "example.md"

s := split.NewDocumentStructureStrategy()
s.ChunkSize = 400
s.MaxDepth = 3
s.SemanticThreshold = 0.5
s.SkipEmptyHeadings = true

chunks, err := s.Split(ctx, text, fileName)
if err != nil {
    panic(err)
}

for _, chunk := range chunks {
    fmt.Println(chunk.SliceContent.Text)
}
```

---

## 策略选择建议

| 场景                 | 推荐策略                   |
| -------------------- | -------------------------- |
| 简单分割，固定长度   | FixedSizeStrategy          |
| 保持句子/段落完整性  | RecursiveCharacterStrategy |
| 语义相似的句子放一起 | SemanticStrategy           |
| Markdown/HTML 文档   | DocumentStructureStrategy  |
