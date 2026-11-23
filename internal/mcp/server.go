// Package mcpserver 把网页生成与浏览器打开能力封装为标准 MCP 工具，
// 通过 stdio 传输对外提供服务（Model Context Protocol）。
package mcpserver

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"agent_w/internal/core"
	"agent_w/internal/shell"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server 持有网页生成 Agent，对外暴露 MCP 工具。
type Server struct {
	webAgent *core.WebAgent
	exec     *shell.Runner
	svr      *mcp.Server
}

// New 创建 MCP Server，webAgent 用于 generate_webpage 工具生成网页。
func New(outDir string, webAgent *core.WebAgent) *Server {
	workDir := filepath.Dir(outDir)
	s := &Server{
		webAgent: webAgent,
		exec:     shell.New(workDir, 30*time.Second),
	}
	s.svr = mcp.NewServer(&mcp.Implementation{Name: "agent-w-mcp", Version: "v1.0.0"}, nil)

	mcp.AddTool(s.svr, &mcp.Tool{
		Name:        "open_html",
		Description: "用系统默认浏览器打开本地 HTML 文件",
	}, s.openHTML)

	mcp.AddTool(s.svr, &mcp.Tool{
		Name:        "generate_webpage",
		Description: "按需求生成静态网页项目并返回项目路径与文件清单",
	}, s.generateWebpage)

	mcp.AddTool(s.svr, &mcp.Tool{
		Name:        "run_command",
		Description: "在项目目录内执行一条 shell 命令，返回标准输出、标准错误与退出码",
	}, s.runCommand)

	return s
}

// ---------- open_html 工具 ----------

// OpenHTMLInput 打开本地 HTML 文件的入参。
type OpenHTMLInput struct {
	Path string `json:"path" jsonschema:"要在浏览器中打开的 index.html 绝对路径"`
}

// OpenHTMLOutput 打开结果。
type OpenHTMLOutput struct {
	Opened  bool   `json:"opened"`
	Message string `json:"message"`
}

func (s *Server) openHTML(_ context.Context, _ *mcp.CallToolRequest, in OpenHTMLInput) (*mcp.CallToolResult, OpenHTMLOutput, error) {
	if err := openBrowser(in.Path); err != nil {
		return nil, OpenHTMLOutput{Opened: false, Message: err.Error()}, err
	}
	return nil, OpenHTMLOutput{Opened: true, Message: "已在默认浏览器打开: " + in.Path}, nil
}

// openBrowser 跨平台调用系统默认浏览器打开文件。
func openBrowser(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

// ---------- generate_webpage 工具 ----------

// GenerateInput 生成网页的入参。
type GenerateInput struct {
	Name  string `json:"name" jsonschema:"项目目录名，英文或拼音"`
	Title string `json:"title" jsonschema:"网页标题"`
	Desc  string `json:"desc" jsonschema:"页面需求描述"`
}

// GenerateOutput 生成结果。
type GenerateOutput struct {
	Root  string   `json:"root"`
	Files []string `json:"files"`
}

func (s *Server) generateWebpage(_ context.Context, _ *mcp.CallToolRequest, in GenerateInput) (*mcp.CallToolResult, GenerateOutput, error) {
	if in.Name == "" {
		in.Name = "site_" + time.Now().Format("20060102150405")
	}
	if in.Title == "" {
		in.Title = in.Name
	}
	p := core.Project{Name: in.Name, Title: in.Title, Desc: in.Desc}
	res, err := s.webAgent.Create(p, nil)
	if err != nil {
		return nil, GenerateOutput{}, err
	}
	return nil, GenerateOutput{Root: res.Root, Files: res.Files}, nil
}

// ---------- run_command 工具 ----------

// RunCommandInput 执行命令的入参。
type RunCommandInput struct {
	Command string `json:"command" jsonschema:"要在项目目录内执行的 shell 命令"`
}

// RunCommandOutput 命令执行结果。
type RunCommandOutput struct {
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Message  string `json:"message"`
}

func (s *Server) runCommand(ctx context.Context, _ *mcp.CallToolRequest, in RunCommandInput) (*mcp.CallToolResult, RunCommandOutput, error) {
	res, err := s.exec.Run(ctx, in.Command)
	if err != nil {
		return nil, RunCommandOutput{Message: err.Error()}, err
	}
	return nil, RunCommandOutput{
		ExitCode: res.ExitCode,
		Stdout:   res.Stdout,
		Stderr:   res.Stderr,
		Message:  fmt.Sprintf("命令退出码 %d", res.ExitCode),
	}, nil
}

// Run 以 stdio 传输启动 MCP Server，阻塞直到客户端断开。
func (s *Server) Run(ctx context.Context) error {
	return s.svr.Run(ctx, &mcp.StdioTransport{})
}
