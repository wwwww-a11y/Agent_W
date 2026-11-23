---
name: open-browser
description: 用系统默认浏览器打开一个本地 HTML 文件
usage: skill open-browser <html 路径>
command: open_html
args: path
triggers: 打开网页, 浏览器, 预览页面
---

# 打开浏览器技能

当用户要求查看、预览某个已经生成好的网页时，调用内置的 `open_html` 动作，
把传入的本地 HTML 绝对路径交给系统默认浏览器打开。

## 参数

- `path`：要打开的 `index.html` 绝对路径。

## 说明

- 仅打开本地文件，不启动本地 HTTP 服务。
- 路径为空时动作会报错并提示补充路径。
