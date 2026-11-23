package core

// FileProvider 负责产出网页文件清单。
// 抽象成接口后，既可以用本地模板实现，也可以替换成大模型生成实现。
// history 为本轮之前的对话上下文，本地模板实现可忽略。
type FileProvider interface {
	Provide(p Project, history []ChatMessage) ([]File, error)
}

// Summarizer 是可选能力：把一批旧对话总结成一段摘要，用于对话压缩。
type Summarizer interface {
	Summarize(msgs []ChatMessage) (string, error)
}

// IntentClassifier 是可选能力：判断输入是否为网页生成需求。
type IntentClassifier interface {
	Classify(text string) (needPage bool, reply string, err error)
}

// LocalProvider 使用内置本地模板生成文件。
type LocalProvider struct{}

// Provide 调用模板生成全部文件。
func (LocalProvider) Provide(p Project, _ []ChatMessage) ([]File, error) {
	return WebFiles(p)
}

// WebAgent 网页生成 Agent。
type WebAgent struct {
	provider FileProvider
	gen      *Generator
}

// New 创建 Agent，outDir 为网页项目输出根目录。
func New(outDir string) *WebAgent {
	return &WebAgent{
		provider: LocalProvider{},
		gen:      NewGenerator(outDir),
	}
}

// WithProvider 替换内容来源（例如接入大模型 Provider）。
func (a *WebAgent) WithProvider(provider FileProvider) *WebAgent {
	a.provider = provider
	return a
}

// Create 根据项目需求生成一个按功能分目录的网页项目。
// history 为会话内已有的对话上下文，可传 nil。
func (a *WebAgent) Create(p Project, history []ChatMessage) (*Result, error) {
	files, err := a.provider.Provide(p, history)
	if err != nil {
		return nil, err
	}
	return a.gen.Build(p, files)
}

// Summarize 委托底层 Provider 总结对话；底层不支持总结时返回空字符串。
func (a *WebAgent) Summarize(msgs []ChatMessage) (string, error) {
	s, ok := a.provider.(Summarizer)
	if !ok {
		return "", nil
	}
	return s.Summarize(msgs)
}

// Classify 委托底层 Provider 判断意图；不支持分类时默认按网页需求处理。
func (a *WebAgent) Classify(text string) (bool, string, error) {
	c, ok := a.provider.(IntentClassifier)
	if !ok {
		return true, "", nil
	}
	return c.Classify(text)
}
