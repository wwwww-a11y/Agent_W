# Agent W — 网页生成 Agent

一个用 Go 写的命令行网页生成助手：输入网页需求后，调用 DeepSeek 大模型生成静态网页文件；没有配置 Key 时回退到内置本地模板。项目同时带本地记忆、Shell 命令执行、声明式技能包、MCP 接入和会话对话压缩。

## 功能特性

- **网页生成**：自然语言描述需求，生成 `index.html`、`css/`、`js/`、`pages/`、`components/` 等静态文件。
- **两种内容来源**：`llm`（DeepSeek，OpenAI 兼容接口）或 `local`（内置模板）。
- **意图判断**：每轮先让模型判断是做页面需求还是闲聊；闲聊只回话，不生成站点。
- **对话历史压缩**：同一会话内保留上下文，每 20 轮把较早的对话压缩成一段摘要。
- **长记忆**：每次生成结果写入本地 `memory/`，跨会话可查看历史摘要。
- **Shell 执行**：限定在项目目录内执行命令，带破坏性命令黑名单。
- **技能包**：`skills/` 目录下的声明式技能，启动时自动扫描加载。
- **MCP 服务**：可作为标准 MCP Server（stdio）对外提供工具。

## 目录结构

```
Agent_W/
├── main.go                      # 入口：REPL 交互循环、new / mcp 子命令
├── go.mod / go.sum              # Go 模块与依赖
├── .env                         # 本地配置（DEEPSEEK_API_KEY，不入库）
├── internal/
│   ├── core/                    # 基础支撑层（单一包）
│   │   ├── model.go             # Project / File / Result 等数据结构
│   │   ├── env.go               # .env 文件读取与环境变量注入
│   │   ├── generator.go          # 创建项目目录并把文件落盘
│   │   ├── template.go           # 内置本地网页模板
│   │   ├── agent.go              # Agent 编排、Provider 接口与可选能力透传
│   │   ├── provider_deepseek.go  # DeepSeek 请求、意图分类、摘要压缩
│   │   └── history.go            # 会话消息历史与每 20 轮压缩
│   ├── mcp/
│   │   ├── server.go            # MCP 工具服务端（open_html / generate_webpage / run_command）
│   │   └── client.go            # 连接本地 stdio MCP 服务的客户端封装
│   ├── memory/
│   │   └── memory.go            # 生成记录持久化与摘要 Markdown
│   ├── shell/
│   │   └── shell.go             # 限定目录内执行命令 + 安全黑名单
│   ├── skills/
│   │   └── skills.go             # 技能目录扫描、解析与触发词匹配
│   └── termui/
│       └── termui.go           # 终端颜色与排版输出（无第三方依赖）
├── skills/                      # 技能源配置（各子目录一个 SKILL.md）
├── memory/                      # 运行时记忆数据（gitignore）
└── output/                      # 生成的网页项目（gitignore）
```

## 运行方式

### 1. 交互式对话（REPL）

```bash
go run .
```

启动后直接输入网页需求即可。可选格式：`项目名:页面需求`（项目名用英文或拼音），例如：

```
my_home:做一个深色个人主页
```

### 2. 一次性命令行生成

```bash
go run . new --name <目录名> [选项]
```

常用选项：

| 选项 | 说明 |
| --- | --- |
| `--name` | 项目目录名（必填，英文或拼音） |
| `--title` | 网页标题（默认同 name） |
| `--desc` | 页面简介 / 需求描述 |
| `--out` | 输出根目录（默认 `output`） |
| `--provider` | `local`（默认，本地模板）或 `llm`（DeepSeek） |
| `--api-key` | DeepSeek Key（优先级高于环境变量和 `.env`） |
| `--model` | 模型名（默认 `deepseek-chat`） |
| `--base-url` | 接口地址（默认 `https://api.deepseek.com`） |

### 3. MCP Server 模式

```bash
go run . mcp
```

以 stdio 方式启动，供外部 MCP 客户端接入。

## 配置

在项目根目录新建 `.env`，写入 DeepSeek Key：

```
DEEPSEEK_API_KEY=你的Key
```

Key 读取优先级：命令行 `--api-key` > 环境变量 `DEEPSEEK_API_KEY` > `.env`。
未配置 Key 时自动回退到本地模板模式。

## REPL 常用指令

| 输入 | 作用 |
| --- | --- |
| `exit` / `quit` / `bye` / `q` / `退出` | 退出程序 |
| `history` / `h` / `记忆` / `记录` | 查看已保存的历史记忆摘要 |
| `skills` | 列出已加载技能 |
| `skill <技能名> [参数]` | 调用指定技能 |
| `!命令` | 直接执行一条 shell 命令，例如 `!go version` |

## MCP 工具

| 工具名 | 作用 |
| --- | --- |
| `open_html` | 用系统默认浏览器打开本地 HTML 文件 |
| `generate_webpage` | 按需求生成静态网页项目，返回路径与文件清单 |
| `run_command` | 在项目目录内执行一条 shell 命令，返回输出与退出码 |

## 技能包

每个技能位于 `skills/<技能名>/SKILL.md`，文件开头用 `---` 包裹元信息：

- `name`：技能名（缺省取目录名）
- `description`：一句话用途
- `usage`：调用方式示例
- `command`：绑定的内置动作，可选 `open_html` / `run_command` / `history` / `generate` / `prompt`
- `triggers`：触发词，逗号分隔

新增技能只需复制一个目录、写好 `SKILL.md`，程序启动时自动加载，无需改动代码。

## 记忆与对话压缩

- **长记忆**：每次成功生成后，把项目名、标题、需求、路径、文件清单记入 `memory/history.json`，并刷新 `memory/summary.md` 摘要。
- **会话压缩**：REPL 内每完成 20 轮对话，调用模型把较早消息总结成一段摘要，之后只携带摘要与近期对话；压缩失败时保留现有历史，不影响生成。

## 其他说明

- `memory/`、`output/`、`*.exe`、`.env` 均已在 `.gitignore` 中忽略。
- 终端颜色在设置 `NO_COLOR` 环境变量、或输出被重定向到文件时自动关闭。
- Go 版本：1.26；唯一外部依赖为 MCP SDK `github.com/modelcontextprotocol/go-sdk`。
