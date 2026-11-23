// Package termui 提供轻量的终端排版与颜色输出。
// 仅使用标准库 ANSI 转义码；设置 NO_COLOR 环境变量、或输出被重定向到文件时，
// 自动关闭颜色，避免出现转义字符。
package termui

import (
	"fmt"
	"os"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	cyan   = "\033[36m"
	gray   = "\033[90m"
)

// enabled 标记颜色是否开启。
var enabled = colorEnabled()

func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	// 输出不是字符终端（如重定向到文件）时不上色。
	if f, err := os.Stdout.Stat(); err == nil {
		if f.Mode()&os.ModeCharDevice == 0 {
			return false
		}
	}
	return true
}

// paint 给文字套上颜色码；颜色关闭时原样返回。
func paint(code, s string) string {
	if !enabled {
		return s
	}
	return code + s + reset
}

// Bold 加粗文字。
func Bold(s string) string { return paint(bold, s) }

// Cyan 青色文字，用于标题、提示符。
func Cyan(s string) string { return paint(cyan, s) }

// Blue 蓝色文字，用于路径、文件名。
func Blue(s string) string { return paint(blue, s) }

// Gray 灰色文字，用于辅助说明。
func Gray(s string) string { return paint(gray, s) }

// Banner 打印欢迎横幅：一行分隔线 + 加粗标题 + 灰色副标题。
func Banner(title, subtitle string) {
	line := "──────────────────────────────"
	fmt.Println(Cyan(line))
	fmt.Println(Bold(Cyan(title)))
	if subtitle != "" {
		fmt.Println(Gray(subtitle))
	}
	fmt.Println(Cyan(line))
}

// Prompt 打印输入提示符。
func Prompt() {
	fmt.Print(Cyan("› "))
}

// Section 打印小节标题。
func Section(title string) {
	fmt.Println(Bold(Cyan(title)))
}

// Success 打印成功结果，绿色标签。
func Success(format string, a ...any) {
	fmt.Println(green + "[成功]" + reset + " " + fmt.Sprintf(format, a...))
}

// Error 打印错误结果，红色标签。
func Error(format string, a ...any) {
	fmt.Println(red + "[失败]" + reset + " " + fmt.Sprintf(format, a...))
}

// Warn 打印注意/提示信息，黄色标签。
func Warn(format string, a ...any) {
	fmt.Println(yellow + "[注意]" + reset + " " + fmt.Sprintf(format, a...))
}

// Info 打印普通进行中的信息。
func Info(format string, a ...any) {
	fmt.Println(Bold("[信息] ") + fmt.Sprintf(format, a...))
}

// Hint 打印一行灰色辅助说明。
func Hint(format string, a ...any) {
	fmt.Println(Gray(fmt.Sprintf(format, a...)))
}

// Item 打印一个列表条目。
func Item(line string) {
	fmt.Println("  - " + line)
}
