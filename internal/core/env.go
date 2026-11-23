package core

import (
	"os"
	"strings"
)

// LoadEnv 依次读取给定路径的 .env 文件并注入环境变量。
// 文件缺失或单行格式错误时静默跳过，不影响主流程。
// 已存在的环境变量优先级高于 .env 文件中的同名键。
func LoadEnv(paths ...string) {
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		parseEnv(data)
	}
}

func parseEnv(data []byte) {
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")

		eq := strings.Index(line, "=")
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		val = strings.Trim(val, `"'`)

		if key == "" {
			continue
		}
		// 外部已设置的环境变量优先，不被 .env 覆盖。
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}
