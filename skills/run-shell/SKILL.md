---
name: run-shell
description: 在项目目录内执行一条 shell 命令，返回输出与退出码
usage: skill run-shell <命令>
command: run_command
args: command
triggers: 执行命令, 运行命令, 终端
---

# 执行 shell 命令技能

在项目根目录内执行用户给定的命令，捕获标准输出、标准错误与退出码。

## 参数

- `command`：要执行的 shell 命令。

## 安全说明

- 命令被限制在项目目录内运行，单条命令超时 30 秒。
- 递归删除、格式化磁盘、改注册表、提权等破坏性命令会被黑名单拦截。
- 不要把该技能用于执行来源不可信的命令。
