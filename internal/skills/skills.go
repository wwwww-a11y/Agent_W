// Package skills 实现声明式技能包（SKILL.md）的扫描与加载。
//
// 每个技能位于 skills/<技能名>/ 目录下，目录内的 SKILL.md 由两部分组成：
//
//	---
//	name: 技能名（必填，缺省时取目录名）
//	description: 一句话描述技能用途
//	usage: 在对话中的调用方式示例
//	command: 绑定的内置动作，可选 open_html / run_command / history / generate / prompt
//	triggers: 触发词，逗号分隔（用于自然语言命中提示）
//	---
//
//	# 技能正文
//	（提示词、使用说明或模板，command=prompt 时作为技能内容直接展示）
//
// 新增技能只需复制一个目录并写好 SKILL.md，程序启动时自动扫描加载，无需改动代码。
package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Skill 描述一个加载好的声明式技能。
type Skill struct {
	// Name 技能唯一标识（小写）。
	Name string
	// Description 一句话用途说明。
	Description string
	// Usage 对话中调用方式示例。
	Usage string
	// Command 绑定的内置动作标识。
	// 可选：open_html / run_command / history / generate / prompt。
	Command string
	// Triggers 自然语言触发词。
	Triggers []string
	// Prompt SKILL.md front-matter 之后的正文（说明/提示词/模板）。
	Prompt string
}

// Registry 是已加载技能的注册表，支持按名查找与触发词匹配。
type Registry struct {
	dir     string
	list    []*Skill
	byName  map[string]*Skill
}

// Load 扫描 dir 下的一级子目录，加载其中的 SKILL.md。
// 目录不存在时返回空注册表（不视为错误），便于首次运行。
func Load(dir string) (*Registry, error) {
	r := &Registry{dir: dir, byName: map[string]*Skill{}}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return r, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillDir := filepath.Join(dir, entry.Name())
		sk, err := loadSkill(skillDir)
		if err != nil {
			// 单个技能解析失败不阻断整体加载，记录后跳过。
			fmt.Fprintf(os.Stderr, "跳过技能 %s: %v\n", entry.Name(), err)
			continue
		}
		if sk == nil {
			continue
		}
		r.list = append(r.list, sk)
		r.byName[sk.Name] = sk
	}
	return r, nil
}

// List 返回全部已加载技能（加载顺序）。
func (r *Registry) List() []*Skill {
	out := make([]*Skill, len(r.list))
	copy(out, r.list)
	return out
}

// Get 按名字查找技能（忽略大小写与首尾空格）。
func (r *Registry) Get(name string) (*Skill, bool) {
	sk, ok := r.byName[strings.ToLower(strings.TrimSpace(name))]
	return sk, ok
}

// Match 在输入文本中按触发词匹配技能；未命中返回 nil。
// 匹配规则：输入包含任一触发词即视为命中，返回第一个登记该触发词的技能。
func (r *Registry) Match(text string) *Skill {
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return nil
	}
	for _, sk := range r.list {
		for _, t := range sk.Triggers {
			t = strings.ToLower(strings.TrimSpace(t))
			if t != "" && strings.Contains(text, t) {
				return sk
			}
		}
	}
	return nil
}

// loadSkill 解析单个技能目录下的 SKILL.md。
// 目录内没有 SKILL.md 时返回 nil, nil（跳过该目录）。
func loadSkill(skillDir string) (*Skill, error) {
	mdPath := filepath.Join(skillDir, "SKILL.md")
	data, err := os.ReadFile(mdPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	raw := string(data)
	front := map[string]string{}
	prompt := strings.TrimSpace(raw)

	// 解析 `---` 包围的 front-matter。
	if strings.HasPrefix(prompt, "---") {
		rest := strings.TrimPrefix(prompt, "---")
		end := strings.Index(rest, "---")
		if end >= 0 {
			meta := rest[:end]
			prompt = strings.TrimSpace(rest[end+3:])
			for _, line := range strings.Split(meta, "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				colon := strings.Index(line, ":")
				if colon < 0 {
					continue
				}
				key := strings.ToLower(strings.TrimSpace(line[:colon]))
				val := strings.TrimSpace(line[colon+1:])
				front[key] = val
			}
		}
	}

	name := front["name"]
	if name == "" {
		name = filepath.Base(skillDir)
	}
	name = strings.ToLower(strings.TrimSpace(name))

	command := front["command"]
	if command == "" {
		command = "prompt"
	}

	skill := &Skill{
		Name:        name,
		Description: front["description"],
		Usage:       front["usage"],
		Command:     command,
		Triggers:    splitCSV(front["triggers"]),
		Prompt:      prompt,
	}
	return skill, nil
}

// splitCSV 把逗号分隔的字符串拆成去空、去重的小写触发词列表。
func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p != "" && !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}
