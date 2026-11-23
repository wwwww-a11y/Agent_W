// Package core 汇总 Agent 的基础支撑能力：数据结构、.env 加载、
// 文件落盘、网页模板、编排层与大模型 Provider。
package core

// Project 描述一个待生成的网页项目。
type Project struct {
	// Name 项目目录名（英文/拼音，避免路径中的特殊字符）。
	Name string
	// Title 网页在浏览器中显示的标题。
	Title string
	// Desc 页面简介，用于首页正文与 meta 描述。
	Desc string
}

// File 表示生成出的单个文件：相对路径 + 文件内容。
type File struct {
	RelPath string
	Content string
}

// Result 记录一次生成的结果。
type Result struct {
	// Root 生成项目的绝对根目录。
	Root string
	// Files 本次写入的文件相对路径列表。
	Files []string
}
