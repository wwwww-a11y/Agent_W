package core

import (
	"os"
	"path/filepath"
)

// Generator 把生成结果写入指定项目根目录。
type Generator struct {
	// base 输出根目录，项目子目录会拼在它下面。
	base string
}

// NewGenerator 创建一个 Generator，base 为网页项目输出根目录。
func NewGenerator(base string) *Generator {
	return &Generator{base: base}
}

// Build 在 base/Name 下创建目录并写入 files，返回生成结果。
func (g *Generator) Build(p Project, files []File) (*Result, error) {
	root := filepath.Join(g.base, p.Name)

	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}

	result := &Result{Root: root}
	for _, f := range files {
		full := filepath.Join(root, f.RelPath)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(full, []byte(f.Content), 0o644); err != nil {
			return nil, err
		}
		result.Files = append(result.Files, f.RelPath)
	}
	return result, nil
}
