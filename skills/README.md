# skills 目录：声明式技能包

每个技能一个子目录，目录内放一份 `SKILL.md`。程序启动时自动扫描本目录加载，
**新增技能无需改动 Go 代码**。

## 目录结构

```
skills/
├── open-browser/        # 打开浏览器
│   └── SKILL.md
├── run-shell/           # 执行 shell 命令
│   └── SKILL.md
├── view-history/       # 查看历史记忆
│   └── SKILL.md
├── make-webpage/       # 生成网页
│   └── SKILL.md
└── landing-page-guide/ # 纯提示词模板技能
    └── SKILL.md
```

## SKILL.md 格式

文件顶部是 `---` 包围的元信息（front-matter），下面是正文（提示词/说明/模板）：

```markdown
---
name: my-skill
description: 一句话说明这个技能做什么
usage: skill my-skill <参数>
command: prompt
args: 参数说明
triggers: 触发词一, 触发词二
---

# 技能正文

这里写提示词、使用说明或模板。command 为 prompt 时，调用技能会直接展示这段正文。
```

### 字段说明

| 字段 | 必填 | 说明 |
|------|------|------|
| `name` | 否 | 技能唯一标识；缺省取目录名，调用时忽略大小写 |
| `description` | 否 | 列表中展示的用途说明 |
| `usage` | 否 | 对话里的调用示例 |
| `command` | 否 | 绑定的内置动作：`open_html` / `run_command` / `history` / `generate` / `prompt`（默认） |
| `args` | 否 | 参数占位说明 |
| `triggers` | 否 | 逗号分隔的触发词，用于自然语言命中提示 |

### 内置动作（command）

| command | 行为 | 需要的参数 |
|---------|------|-----------|
| `open_html` | 调用 MCP `open_html` 在浏览器打开本地 HTML | HTML 路径 |
| `run_command` | 在项目目录执行 shell 命令 | 命令文本 |
| `history` | 输出本地记忆摘要 | 无 |
| `generate` | 按需求走网页生成流程 | 需求描述 |
| `prompt` | 仅展示 SKILL.md 正文（纯提示词/模板） | 无 |

## 如何新增一个技能

1. 在本目录新建子目录，如 `skills/my-skill/`。
2. 放入 `SKILL.md`，按上面格式填写元信息与正文。
3. 若复用内置动作，把 `command` 填成上表之一；纯说明类留 `prompt`。
4. 重启程序，对话输入 `skills` 即可看到新技能。

> 若需要全新的内置动作类型（现有五种之外），才需要在代码里补充对应的动作分发逻辑。
