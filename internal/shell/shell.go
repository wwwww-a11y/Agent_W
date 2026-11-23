// Package shell 在限定工作目录内执行 shell 命令，
// 并对破坏性、系统级命令做黑名单拦截。
package shell

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Result 一次命令执行的结果。
type Result struct {
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

// Runner 持有命令执行的工作目录与超时设置。
type Runner struct {
	workDir string
	timeout time.Duration
}

// New 创建执行器，workDir 为命令允许运行的工作目录，timeout 为单条命令超时。
func New(workDir string, timeout time.Duration) *Runner {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Runner{workDir: workDir, timeout: timeout}
}

// Run 执行一条 shell 命令，返回输出与退出码。
// 命令被安全策略拦截、或进程启动失败时返回 error；退出码非 0 仍作为 Result 返回。
func (r *Runner) Run(ctx context.Context, command string) (*Result, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return nil, fmt.Errorf("命令为空")
	}

	if reason := blockedReason(command); reason != "" {
		return nil, fmt.Errorf("命令被安全策略拦截: %s", reason)
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	cmd.Dir = r.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	res := &Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if cmd.ProcessState != nil {
		res.ExitCode = cmd.ProcessState.ExitCode()
	}

	if err != nil {
		// 退出码非 0 属于正常执行结果，不按致命错误返回；仅启动失败才返回 error。
		if _, ok := err.(*exec.ExitError); !ok {
			return res, err
		}
	}
	return res, nil
}

// blockedReason 返回拦截原因；空字符串表示放行。
func blockedReason(command string) string {
	lower := strings.ToLower(command)

	for _, item := range blocked {
		if strings.Contains(lower, item.pattern) {
			return item.reason
		}
	}
	return ""
}

type blockedItem struct {
	pattern string
	reason  string
}

// blocked 为破坏性 / 系统级命令黑名单。常见开发命令（go、dir、ls、
// mkdir、cp、move、type、echo 等）不在拦截范围内。
var blocked = []blockedItem{
	{"rm -rf", "递归强制删除"},
	{"rm -r", "递归删除"},
	{"rmdir", "删除目录"},
	{"del ", "删除文件"},
	{"erase ", "删除文件"},
	{"format ", "磁盘格式化"},
	{"mkfs", "创建文件系统"},
	{"fdisk", "磁盘分区"},
	{"dd ", "磁盘/设备写入"},
	{"shutdown", "关机"},
	{"reboot", "重启"},
	{"poweroff", "关机"},
	{"halt", "停机"},
	{"reg add", "修改注册表"},
	{"reg delete", "删除注册表项"},
	{"net user", "修改系统用户"},
	{"taskkill", "强制结束进程"},
	{"chmod", "修改权限"},
	{"chown", "修改属主"},
	{"sudo ", "提权执行"},
	{"mkfifo", "创建管道"},
	{"mount ", "挂载文件系统"},
	{"umount", "卸载文件系统"},
	{"diskpart", "磁盘分区操作"},
	{"bcdedit", "修改启动配置"},
	{"icacls", "修改访问控制列表"},
	{"attrib ", "修改文件属性"},
	{"curl ", "网络下载后执行存在风险，已拦截"},
	{"wget ", "网络下载后执行存在风险，已拦截"},
	{":(){", "fork 炸弹"},
	{">/dev/sd", "直接写磁盘设备"},
}
