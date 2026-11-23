package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DeepSeekProvider 调用 DeepSeek 大模型生成网页文件清单。
// 接口兼容 OpenAI Chat Completions 格式。
type DeepSeekProvider struct {
	APIKey    string
	Model     string
	BaseURL   string
	MaxTokens int
	Client    *http.Client
}

// NewDeepSeekProvider 填充默认值并创建 Provider。
// apiKey 为空时调用方应提前校验。
func NewDeepSeekProvider(apiKey, modelName, baseURL string) *DeepSeekProvider {
	if modelName == "" {
		modelName = "deepseek-chat"
	}
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	return &DeepSeekProvider{
		APIKey:    apiKey,
		Model:     modelName,
		BaseURL:   strings.TrimRight(baseURL, "/"),
		MaxTokens: 32768,
		Client:    &http.Client{Timeout: 120 * time.Second},
	}
}

// chatMessage DeepSeek/OpenAI 兼容消息结构。
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message      chatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// llmFile 是模型返回的单个文件条目。
type llmFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type llmResult struct {
	Files []llmFile `json:"files"`
}

// Provide 调用 DeepSeek 生成网页文件，并按功能目录整理为 File。
// history 为会话内已有对话上下文；解析失败或输出被截断时重试一次。
func (d *DeepSeekProvider) Provide(p Project, history []ChatMessage) ([]File, error) {
	var lastErr error
	for attempt := 1; attempt <= 2; attempt++ {
		content, finish, err := d.generateOnce(p, history, attempt > 1)
		if err != nil {
			// 网络、鉴权、协议类错误不重试。
			return nil, err
		}

		if finish == "length" {
			lastErr = fmt.Errorf("模型输出达到长度上限被截断（finish_reason=length），页面代码过长，请精简需求后重试")
			continue
		}

		files, perr := parseModelFiles(content)
		if perr == nil {
			return files, nil
		}
		lastErr = perr
	}
	return nil, lastErr
}

// Summarize 把一批旧对话压缩成一段简短中文摘要，保留需求、方案与已生成项目。
func (d *DeepSeekProvider) Summarize(msgs []ChatMessage) (string, error) {
	var b strings.Builder
	for _, m := range msgs {
		fmt.Fprintf(&b, "- %s: %s\n", m.Role, m.Content)
	}
	messages := []chatMessage{
		{Role: "system", Content: "你在帮一个网页生成 Agent 整理历史对话。请把下面的对话压缩成一段简短的中文摘要，保留：用户提过的需求、确定过的方案、已经生成过哪些项目。不要罗列原始代码，不要输出 JSON，只输出摘要本身。"},
		{Role: "user", Content: b.String()},
	}
	content, _, err := d.chat(messages)
	return strings.TrimSpace(content), err
}

// classifyResult 意图判断的返回结构。
type classifyResult struct {
	NeedPage bool   `json:"need_page"`
	Reply    string `json:"reply"`
}

// Classify 判断用户输入是网页生成需求，还是闲聊/问候/其他。
// 返回 needPage：是否需要生成页面；reply：闲聊时回应用户的话术。
// 请求或解析失败时按"需要生成"兜底，避免误伤正常需求。
func (d *DeepSeekProvider) Classify(text string) (needPage bool, reply string, err error) {
	messages := []chatMessage{
		{Role: "system", Content: `你是网页生成助手的意图判断模块。判断下面这条用户输入属于哪一类：
- need_page=true：用户想制作网站、网页、主页、介绍页等，哪怕需求描述不完整。
- need_page=false：问候、打招呼、闲聊、提问、感慨，或与做网页无关的内容。
只输出一个 JSON 对象：{"need_page": true 或 false, "reply": "need_page=false 时给一句简短中文回应；true 时留空"}。不要输出其他内容。`},
		{Role: "user", Content: text},
	}

	content, _, err := d.chat(messages)
	if err != nil {
		return true, "", err
	}

	var res classifyResult
	if perr := json.Unmarshal([]byte(extractJSON(content)), &res); perr != nil {
		return true, "", nil
	}
	return res.NeedPage, strings.TrimSpace(res.Reply), nil
}

// generateOnce 发起一次聊天补全，返回模型文本与停止原因。
func (d *DeepSeekProvider) generateOnce(p Project, history []ChatMessage, retry bool) (string, string, error) {
	messages := []chatMessage{
		{Role: "system", Content: systemPrompt},
	}
	for _, m := range history {
		messages = append(messages, chatMessage{Role: m.Role, Content: m.Content})
	}
	messages = append(messages, chatMessage{Role: "user", Content: buildUserPrompt(p)})
	if retry {
		messages = append(messages, chatMessage{
			Role:    "user",
			Content: "上一次输出不完整。请进一步精简代码、合并样式与脚本，只保留最核心的文件，确保 JSON 完整闭合。",
		})
	}

	return d.chat(messages)
}

// chat 按给定消息发起一次聊天补全，返回模型文本与停止原因。
func (d *DeepSeekProvider) chat(messages []chatMessage) (string, string, error) {
	reqBody := chatRequest{
		Model:     d.Model,
		MaxTokens: d.MaxTokens,
		Messages:  messages,
	}

	raw, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest(http.MethodPost, d.BaseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.APIKey)

	resp, err := d.Client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("请求 DeepSeek 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("DeepSeek 返回状态 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", "", fmt.Errorf("解析响应失败: %w", err)
	}
	if chatResp.Error != nil {
		return "", "", fmt.Errorf("DeepSeek 错误: %s", chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 {
		return "", "", fmt.Errorf("DeepSeek 未返回任何内容")
	}

	choice := chatResp.Choices[0]
	return choice.Message.Content, choice.FinishReason, nil
}

// parseModelFiles 从模型文本中提取 JSON 文件清单，并做路径安全校验。
func parseModelFiles(content string) ([]File, error) {
	jsonStr := extractJSON(content)
	var result llmResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("解析文件清单失败: %w（若内容在中途断开，通常是输出被长度截断；原始内容：%s）", err, truncate(content, 300))
	}

	files := make([]File, 0, len(result.Files))
	for _, f := range result.Files {
		rel, err := safeRelPath(f.Path)
		if err != nil {
			return nil, err
		}
		files = append(files, File{RelPath: rel, Content: f.Content})
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("模型未生成任何文件")
	}
	return files, nil
}

// safeRelPath 校验相对路径，防止目录逃逸，并限制在允许的功能目录内。
func safeRelPath(path string) (string, error) {
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("存在空路径的文件")
	}
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part == ".." || part == "." {
			return "", fmt.Errorf("非法路径: %s", path)
		}
	}

	allowedRoot := map[string]bool{"index.html": true, "README.md": true}
	allowedPrefix := []string{"pages/", "components/", "css/", "js/", "assets/"}

	if allowedRoot[path] {
		return path, nil
	}
	for _, prefix := range allowedPrefix {
		if strings.HasPrefix(path, prefix) {
			return path, nil
		}
	}
	return "", fmt.Errorf("路径不在允许的功能目录内: %s", path)
}

// extractJSON 剥离可能存在的 ```json 代码块包裹，取出首个 JSON 对象。
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if start := strings.Index(s, "```"); start >= 0 {
		rest := s[start+3:]
		rest = strings.TrimPrefix(rest, "json")
		rest = strings.TrimPrefix(rest, "JSON")
		if end := strings.LastIndex(rest, "```"); end >= 0 {
			rest = rest[:end]
		}
		s = strings.TrimSpace(rest)
	}
	first := strings.Index(s, "{")
	last := strings.LastIndex(s, "}")
	if first >= 0 && last > first {
		return s[first : last+1]
	}
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

const systemPrompt = `你是一名资深前端工程师。根据用户给出的站点需求，生成一个完整、可直接用浏览器打开的静态网站。

严格要求：
1. 只输出一个 JSON 对象，不要任何解释、不要 Markdown 代码块包裹。
2. JSON 结构必须为：
{"files":[{"path":"index.html","content":"文件完整内容"}, ...]}
3. 文件必须按功能放到以下目录，path 使用正斜杠：
   - index.html            首页入口
   - pages/                 子页面，如 pages/about.html
   - components/            可复用 HTML 片段，如 components/header.html
   - css/                   样式表，如 css/main.css
   - js/                    脚本，如 js/main.js
   - assets/                静态资源说明
4. HTML 中引用 css、js 时使用相对路径；页面要能直接打开运行。
5. content 必须是文件的完整内容，JSON 字符串内的换行、引号要正确转义。
6. 代码务必精简：少写注释和冗余样式，优先只保留 index.html、css/main.css、js/main.js 三个核心文件；子页面和组件按需添加，不要堆砌文件。
7. 输出结束前确认整个 JSON 对象已完整闭合（括号、引号、逗号齐全）。`

func buildUserPrompt(p Project) string {
	return fmt.Sprintf(`请生成网站：
- 项目名：%s
- 网页标题：%s
- 站点简介/需求：%s

请围绕上述需求设计页面内容、样式与交互，并输出符合约定 JSON 结构的文件清单。`,
		p.Name, p.Title, p.Desc)
}
