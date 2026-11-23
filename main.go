// Agent 命令行入口。
// 直接运行进入交互式对话（REPL）；`new` 子命令用于一次性命令行生成。
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"agent_w/internal/core"
	mcpserver "agent_w/internal/mcp"
	"agent_w/internal/memory"
	"agent_w/internal/skills"
	"agent_w/internal/shell"
	"agent_w/internal/termui"
)

func main() {
	loadDotEnv()

	// 带子命令时走一次性命令；否则进入交互式 Agent。
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "new":
			create()
			return
		case "mcp":
			runMCPServer()
			return
		case "help", "-h", "--help":
			usage()
			return
		}
	}
	runRepl()
}

// runRepl 交互式循环：读取用户需求，生成网页，直到输入退出指令。
func runRepl() {
	outDir, err := filepath.Abs("output")
	if err != nil {
		log.Fatal(err)
	}

	// 有 Key 用 DeepSeek，否则回退本地模板。
	providerKind := defaultProvider()

	a, err := buildAgent(outDir, providerKind, "", "", "")
	if err != nil {
		termui.Warn("未启用大模型，改用本地模板：%v", err)
		a, _ = buildAgent(outDir, "local", "", "", "")
	}

	mem, err := memory.New(memoryDir())
	if err != nil {
		termui.Warn("记忆初始化失败（不影响生成）：%v", err)
		mem = nil
	}

	workDir, _ := filepath.Abs(".")
	runner := shell.New(workDir, 30*time.Second)

	reg, err := skills.Load(skillsDir())
	if err != nil {
		termui.Warn("技能加载失败（不影响生成）：%v", err)
		reg, _ = skills.Load("")
	}

	// 会话内对话历史：每 20 轮把较早的对话压缩成一段摘要。
	hist := core.NewChatHistory(20)

	termui.Banner("网页生成 Agent", "来源："+providerKind)
	termui.Hint("直接描述你想要的网页；输入 exit / quit 退出；输入 history 查看历史摘要。")
	termui.Hint("输入 ! 开头可直接执行 shell 命令，例如 !go version")
	termui.Hint("输入 skills 查看已加载技能；输入 skill <技能名> [参数] 调用技能。")
	termui.Hint("可选格式：项目名:页面需求（项目名用英文或拼音），例如 my_home:做一个深色个人主页")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println()
		termui.Prompt()
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if isExit(input) {
			termui.Info("再见！")
			break
		}
		if isHistory(input) {
			if mem != nil {
				text, _ := mem.SummaryText()
				termui.Section("历史记忆")
				fmt.Println(text)
			}
			continue
		}

		if isSkillsList(input) {
			printSkills(reg)
			continue
		}
		if skillName, arg, ok := parseSkillCall(input); ok {
			runSkill(reg, skillName, arg, a, mem, scanner, providerKind, runner, hist)
			continue
		}

		if strings.HasPrefix(input, "!") {
			cmd := strings.TrimSpace(strings.TrimPrefix(input, "!"))
			res, err := runner.Run(context.Background(), cmd)
			if err != nil {
				termui.Error("执行失败：%v", err)
			} else {
				if res.Stdout != "" {
					fmt.Print(res.Stdout)
				}
				if res.Stderr != "" {
					termui.Hint("stderr: %s", res.Stderr)
				}
				termui.Hint("退出码: %d", res.ExitCode)
			}
			continue
		}

		// 自然语言输入：先让模型判断是做页面需求，还是闲聊/其他。
		needPage, reply, err := a.Classify(input)
		if err != nil {
			termui.Warn("意图判断失败，按网页需求处理：%v", err)
		} else if !needPage {
			if reply != "" {
				fmt.Println(reply)
			} else {
				termui.Hint("你好，想做个什么网页？直接描述需求即可。")
			}
			continue
		}

		// 若命中技能触发词，先提示可调用的技能（不拦截生成流程）。
		if m := reg.Match(input); m != nil {
			termui.Hint("可使用技能 skill %s", m.Name)
		}

		runGeneration(a, mem, scanner, providerKind, input, hist)
	}
}

// runMCPServer 以标准 MCP Server 模式运行（stdio），供外部 MCP 客户端接入。
func runMCPServer() {
	outDir, err := filepath.Abs("output")
	if err != nil {
		log.Fatal(err)
	}
	a, err := buildAgent(outDir, defaultProvider(), "", "", "")
	if err != nil {
		log.Fatal(err)
	}
	srv := mcpserver.New(outDir, a)
	if err := srv.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

// defaultProvider 根据是否配置 Key 选择默认内容来源。
func defaultProvider() string {
	if os.Getenv("DEEPSEEK_API_KEY") != "" {
		return "llm"
	}
	return "local"
}

// askOpen 询问用户是否在浏览器打开，默认不打开。
func askOpen(scanner *bufio.Scanner) bool {
	fmt.Print(termui.Bold("是否在浏览器打开？(y/N) "))
	if !scanner.Scan() {
		return false
	}
	ans := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return ans == "y" || ans == "yes" || ans == "是"
}

// openViaMCP 通过本地 stdio MCP 客户端调用 open_html 工具打开页面。
func openViaMCP(htmlPath string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	ctx := context.Background()
	client, err := mcpserver.ConnectToCommand(ctx, exe, "mcp")
	if err != nil {
		return err
	}
	defer client.Close()
	return client.OpenHTML(ctx, htmlPath)
}

// create 一次性命令行生成（保留原有 CLI 用法）。
func create() {
	fs := flag.NewFlagSet("new", flag.ExitOnError)
	name := fs.String("name", "", "项目目录名（必填，英文或拼音）")
	title := fs.String("title", "", "网页标题（默认同 name）")
	desc := fs.String("desc", "一个由 Agent 生成的网页项目。", "页面简介/需求描述")
	out := fs.String("out", "output", "网页项目输出根目录")
	provider := fs.String("provider", "local", "内容来源：local=本地模板，llm=DeepSeek 生成")
	apiKey := fs.String("api-key", "", "DeepSeek API Key（参数 > 环境变量 DEEPSEEK_API_KEY > .env）")
	modelName := fs.String("model", "", "模型名，默认 deepseek-chat")
	baseURL := fs.String("base-url", "", "接口地址，默认 https://api.deepseek.com")
	_ = fs.Parse(os.Args[2:])

	if *name == "" {
		log.Fatal("缺少必填参数 --name")
	}
	if *title == "" {
		*title = *name
	}

	outDir, err := filepath.Abs(*out)
	if err != nil {
		log.Fatal(err)
	}

	a, err := buildAgent(outDir, *provider, *apiKey, *modelName, *baseURL)
	if err != nil {
		log.Fatal(err)
	}

	p := core.Project{
		Name:  *name,
		Title: *title,
		Desc:  *desc,
	}

	result, err := a.Create(p, nil)
	if err != nil {
		log.Fatalf("生成失败: %v", err)
	}

	termui.Section("生成完成")
	termui.Item("项目根目录: " + termui.Blue(result.Root))
	termui.Section("文件清单")
	for _, f := range result.Files {
		termui.Item(termui.Blue(f))
	}

	if mem, err := memory.New(memoryDir()); err == nil {
		record(mem, p, result, *provider)
	}
}

// buildAgent 按 provider 构造 Agent，统一处理 Key 读取。
func buildAgent(outDir, provider, apiKey, modelName, baseURL string) (*core.WebAgent, error) {
	a := core.New(outDir)
	switch provider {
	case "local":
		// 内置本地模板，无需 Key。
	case "llm":
		key := apiKey
		if key == "" {
			key = os.Getenv("DEEPSEEK_API_KEY")
		}
		if key == "" {
			return nil, fmt.Errorf("未配置 DeepSeek Key：请在 .env 或环境变量 DEEPSEEK_API_KEY 中设置，或改用 local")
		}
		a = a.WithProvider(core.NewDeepSeekProvider(key, modelName, baseURL))
	default:
		return nil, fmt.Errorf("未知 provider: %s（可选 local / llm）", provider)
	}
	return a, nil
}

// splitRequest 解析 "项目名:需求" 或 "项目名|需求" 格式。
// 项目名含中文时不作为名称，整句视为需求。
func splitRequest(input string) (name, desc string) {
	for _, sep := range []string{":", "：", "|"} {
		if i := strings.Index(input, sep); i >= 0 {
			head := strings.TrimSpace(input[:i])
			tail := strings.TrimSpace(input[i+1:])
			if head != "" && !hasCJK(head) {
				return head, tail
			}
			break
		}
	}
	return "", input
}

func hasCJK(s string) bool {
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}

func autoName() string {
	return "site_" + time.Now().Format("20060102150405")
}

func titleFromDesc(desc string) string {
	t := strings.TrimSpace(desc)
	if t == "" {
		return "未命名站点"
	}
	r := []rune(t)
	if len(r) > 24 {
		r = r[:24]
	}
	return string(r)
}

func isExit(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "exit", "quit", "bye", "q", "退出":
		return true
	}
	return false
}

// isHistory 判断是否为查看历史摘要的命令。
func isHistory(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "history", "h", "记忆", "记录":
		return true
	}
	return false
}

// record 写入一条生成记忆；mem 为 nil 时跳过。
func record(mem *memory.Store, p core.Project, res *core.Result, provider string) {
	if mem == nil {
		return
	}
	_ = mem.Add(memory.Entry{
		Name:     p.Name,
		Title:    p.Title,
		Desc:     p.Desc,
		Provider: provider,
		Root:     res.Root,
		Files:    res.Files,
	})
}

// memoryDir 返回记忆目录的绝对路径。
func memoryDir() string {
	dir, _ := filepath.Abs("memory")
	return dir
}

// skillsDir 返回技能包目录的绝对路径。
func skillsDir() string {
	dir, _ := filepath.Abs("skills")
	return dir
}

// runGeneration 按用户需求生成网页、记录记忆并按需打开浏览器。
// hist 为会话内对话历史，用于携带上下文并在达到阈值后压缩。
func runGeneration(a *core.WebAgent, mem *memory.Store, scanner *bufio.Scanner, providerKind, input string, hist *core.ChatHistory) {
	name, desc := splitRequest(input)
	if name == "" {
		name = autoName()
	}
	title := titleFromDesc(desc)

	p := core.Project{
		Name:  name,
		Title: title,
		Desc:  desc,
	}

	hist.AddUser(desc)

	termui.Info("生成中，请稍候...")
	result, err := a.Create(p, hist.Context())
	if err != nil {
		termui.Error("生成失败：%v", err)
		return
	}

	termui.Success("生成完成：%s", termui.Blue(result.Root))
	termui.Section("文件清单")
	for _, f := range result.Files {
		termui.Item(termui.Blue(f))
	}

	hist.AddAssistant(assistantSummary(p, result))
	record(mem, p, result, providerKind)

	// 攒满阈值后，把较早的对话压缩成一段摘要。
	if hist.ShouldCompress() {
		summary, err := a.Summarize(hist.Recent())
		if err != nil {
			termui.Warn("对话压缩失败，保留现有历史：%v", err)
		} else if summary != "" {
			hist.Compress(summary)
			termui.Hint("已把较早的对话压缩成摘要")
		}
	}

	if askOpen(scanner) {
		indexPath := filepath.Join(result.Root, "index.html")
		if err := openViaMCP(indexPath); err != nil {
			termui.Error("打开失败：%v", err)
		} else {
			termui.Success("已通过 MCP 工具在浏览器打开。")
		}
	}
}

// assistantSummary 把一次生成结果压成简短的对话回复，记入历史。
func assistantSummary(p core.Project, res *core.Result) string {
	return fmt.Sprintf("生成项目 %s（标题：%s），文件：%s", p.Name, p.Title, strings.Join(res.Files, ", "))
}

// isSkillsList 判断是否为列出全部技能的命令。
func isSkillsList(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "skills", "skill list", "skill ls", "技能":
		return true
	}
	return false
}

// parseSkillCall 解析 "skill <名字> [参数]" 形式的调用。
func parseSkillCall(input string) (name, arg string, ok bool) {
	rest := strings.TrimSpace(input)
	if !strings.HasPrefix(strings.ToLower(rest), "skill") {
		return "", "", false
	}
	rest = strings.TrimSpace(rest[len("skill"):])
	if rest == "" {
		return "", "", false
	}
	fields := strings.Fields(rest)
	name = fields[0]
	arg = strings.TrimSpace(rest[len(fields[0]):])
	return name, arg, true
}

// printSkills 列出已加载技能及其用途。
func printSkills(reg *skills.Registry) {
	list := reg.List()
	termui.Section(fmt.Sprintf("已加载技能 %d 个", len(list)))
	for _, sk := range list {
		line := termui.Cyan(sk.Name) + " — " + sk.Description
		if sk.Usage != "" {
			line += "（" + sk.Usage + "）"
		}
		termui.Item(line)
	}
}

// runSkill 按技能声明的内置动作分发执行。
func runSkill(reg *skills.Registry, name, arg string, a *core.WebAgent, mem *memory.Store, scanner *bufio.Scanner, providerKind string, runner *shell.Runner, hist *core.ChatHistory) {
	sk, ok := reg.Get(name)
	if !ok {
		termui.Error("未找到技能：%s（输入 skills 查看可用技能）", name)
		return
	}

	switch sk.Command {
	case "history":
		if mem != nil {
			text, _ := mem.SummaryText()
			termui.Section("历史记忆")
			fmt.Println(text)
		} else {
			termui.Hint("记忆未初始化。")
		}
	case "open_html":
		if arg == "" {
			termui.Error("缺少参数：HTML 文件路径（%s）", sk.Usage)
			return
		}
		if err := openViaMCP(arg); err != nil {
			termui.Error("打开失败：%v", err)
		} else {
			termui.Success("已通过 MCP 工具在浏览器打开。")
		}
	case "run_command":
		if arg == "" {
			termui.Error("缺少参数：要执行的命令（%s）", sk.Usage)
			return
		}
		res, err := runner.Run(context.Background(), arg)
		if err != nil {
			termui.Error("执行失败：%v", err)
		} else {
			if res.Stdout != "" {
				fmt.Print(res.Stdout)
			}
			if res.Stderr != "" {
				termui.Hint("stderr: %s", res.Stderr)
			}
			termui.Hint("退出码: %d", res.ExitCode)
		}
	case "generate":
		if arg == "" {
			termui.Error("缺少参数：页面需求（%s）", sk.Usage)
			return
		}
		runGeneration(a, mem, scanner, providerKind, arg, hist)
	case "prompt":
		fmt.Println(sk.Prompt)
	default:
		termui.Warn("技能 %s 绑定了未知动作: %s", sk.Name, sk.Command)
	}
}

// loadDotEnv 在工作目录、exe 目录及其上级目录查找 .env 并加载。
func loadDotEnv() {
	candidates := []string{".env"}
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, ".env"),
			filepath.Join(exeDir, "..", ".env"),
		)
	}
	core.LoadEnv(candidates...)
}

func usage() {
	termui.Section("用法")
	termui.Hint("  go run .                  进入交互式对话，直接输入网页需求")
	termui.Hint("  go run . new --name <目录名> [选项]   一次性命令行生成")
	termui.Hint("  go run . mcp              以 MCP Server 模式运行（stdio，供外部客户端接入）")
	termui.Hint("  对话中输入 history        查看已保存的历史记忆摘要")
	termui.Hint("  对话中输入 skills         查看已加载的声明式技能")
	termui.Hint("  对话中输入 skill <名>     调用技能，如 skill open-browser <路径>")
	fmt.Println()
	termui.Section("选项（new 子命令）")
	termui.Hint("  --title     网页标题（默认同 name）")
	termui.Hint("  --desc      页面简介/需求描述")
	termui.Hint("  --out       输出根目录（默认 output）")
	termui.Hint("  --provider  local（默认，本地模板）| llm（DeepSeek 生成）")
	termui.Hint("  --api-key   DeepSeek API Key（也可用环境变量 DEEPSEEK_API_KEY）")
	termui.Hint("  --model     模型名（默认 deepseek-chat）")
	termui.Hint("  --base-url  接口地址（默认 https://api.deepseek.com）")
}
