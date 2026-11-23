// Client 是连接到本地 stdio MCP Server 的客户端封装：建立会话、调用 open_html。
package mcpserver

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Client 是连接到本地 stdio MCP Server 的客户端封装。
type Client struct {
	session *mcp.ClientSession
}

// ConnectToCommand 启动指定命令作为 MCP Server 并建立会话。
// command 为可执行文件路径，args 为启动参数（如 ["mcp"]）。
func ConnectToCommand(ctx context.Context, command string, args ...string) (*Client, error) {
	client := mcp.NewClient(
		&mcp.Implementation{Name: "agent-w-client", Version: "v1.0.0"},
		nil,
	)
	transport := &mcp.CommandTransport{Command: exec.Command(command, args...)}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, err
	}
	return &Client{session: session}, nil
}

// OpenHTML 通过 MCP 工具 open_html 打开本地 HTML 文件。
func (c *Client) OpenHTML(ctx context.Context, path string) error {
	res, err := c.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "open_html",
		Arguments: map[string]any{"path": path},
	})
	if err != nil {
		return err
	}
	if res.IsError {
		return fmt.Errorf("open_html 工具返回错误")
	}
	return nil
}

// Close 关闭 MCP 会话。
func (c *Client) Close() error {
	if c.session == nil {
		return nil
	}
	return c.session.Close()
}
