---
name: make-webpage
description: 按自然语言需求生成一个静态网页项目
usage: skill make-webpage [项目名:]页面需求
command: generate
args: desc
triggers: 生成网页, 做一个页面, 建站, 写网页
---

# 生成网页技能

把用户的页面需求交给 Agent 生成静态站点，文件按 `index.html`、`pages/`、
`components/`、`css/`、`js/`、`assets/` 分目录落盘。

## 参数

- `desc`：页面需求描述；可选 `项目名:需求` 格式指定英文/拼音目录名。

## 流程

1. 解析项目名与需求（缺省按时间戳自动命名）。
2. 有 DeepSeek Key 时走大模型生成，否则回退本地模板。
3. 生成成功后写入本地记忆，并询问是否在浏览器打开。
