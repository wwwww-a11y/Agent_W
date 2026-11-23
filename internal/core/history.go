package core

import "strings"

// ChatMessage 一条对话消息，用于把上下文传给大模型。
type ChatMessage struct {
	Role    string // user / assistant / system
	Content string
}

// ChatHistory 在一次会话内保存对话；积累到阈值后，把较早的消息压缩成一段摘要。
type ChatHistory struct {
	messages []ChatMessage // 近期还没被压缩的消息
	summary  string        // 已经压缩好的旧对话摘要
	turns    int           // 自上次压缩后完成的轮数
	limit    int           // 每隔多少轮压缩一次
}

// NewChatHistory 创建会话历史；limit 为压缩轮数阈值（<=0 时取 20）。
func NewChatHistory(limit int) *ChatHistory {
	if limit <= 0 {
		limit = 20
	}
	return &ChatHistory{limit: limit}
}

// AddUser 记录一条用户消息。
func (h *ChatHistory) AddUser(content string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}
	h.messages = append(h.messages, ChatMessage{Role: "user", Content: content})
}

// AddAssistant 记录一条模型回复，并完成一轮计数。
func (h *ChatHistory) AddAssistant(content string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}
	h.messages = append(h.messages, ChatMessage{Role: "assistant", Content: content})
	h.turns++
}

// ShouldCompress 是否达到压缩阈值。
func (h *ChatHistory) ShouldCompress() bool {
	return h.turns >= h.limit
}

// Recent 返回当前尚未压缩的消息副本（供总结使用）。
func (h *ChatHistory) Recent() []ChatMessage {
	out := make([]ChatMessage, len(h.messages))
	copy(out, h.messages)
	return out
}

// Compress 用新的摘要替换旧消息：清空近期原话，只保留摘要。
func (h *ChatHistory) Compress(summary string) {
	summary = strings.TrimSpace(summary)
	if summary != "" {
		h.summary = summary
	}
	h.messages = nil
	h.turns = 0
}

// Context 组装发给模型的上下文：旧对话摘要在前，近期原话在后。
func (h *ChatHistory) Context() []ChatMessage {
	out := make([]ChatMessage, 0, len(h.messages)+1)
	if h.summary != "" {
		out = append(out, ChatMessage{
			Role:    "system",
			Content: "以下是此前对话的摘要，请在其基础上继续：\n" + h.summary,
		})
	}
	out = append(out, h.messages...)
	return out
}
