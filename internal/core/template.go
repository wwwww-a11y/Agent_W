package core

import (
	"bytes"
	"fmt"
	tmpl "text/template"
)

// WebFiles 根据项目信息生成全部网页文件清单。
// 模板按功能拆分：页面(pages)、复用组件(components)、样式(css)、
// 脚本(js)、静态资源(assets)。
func WebFiles(p Project) ([]File, error) {
	raw := map[string]string{
		"index.html":               indexHTML,
		"pages/about.html":         aboutHTML,
		"components/header.html":    headerHTML,
		"components/footer.html":    footerHTML,
		"css/main.css":             mainCSS,
		"js/main.js":               mainJS,
		"assets/README.md":         assetsReadme,
		"README.md":                projectReadme,
	}

	files := make([]File, 0, len(raw))
	for rel, body := range raw {
		content, err := render(body, p)
		if err != nil {
			return nil, fmt.Errorf("渲染 %s 失败: %w", rel, err)
		}
		files = append(files, File{RelPath: rel, Content: content})
	}
	return files, nil
}

// render 用项目数据替换模板占位符。
func render(body string, p Project) (string, error) {
	tpl, err := tmpl.New("web").Parse(body)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, p); err != nil {
		return "", err
	}
	return buf.String(), nil
}

const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="description" content="{{.Desc}}">
  <title>{{.Title}} - 首页</title>
  <link rel="stylesheet" href="css/main.css">
</head>
<body>
  <header class="site-header">
    <h1>{{.Title}}</h1>
    <nav>
      <a href="index.html">首页</a>
      <a href="pages/about.html">关于</a>
    </nav>
  </header>

  <main class="container">
    <section class="hero">
      <h2>欢迎来到 {{.Title}}</h2>
      <p>{{.Desc}}</p>
    </section>
  </main>

  <footer class="site-footer">
    <p>&copy; {{.Title}}. 保留所有权利。</p>
  </footer>

  <script src="js/main.js"></script>
</body>
</html>
`

const aboutHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.Title}} - 关于</title>
  <link rel="stylesheet" href="../css/main.css">
</head>
<body>
  <header class="site-header">
    <h1>{{.Title}}</h1>
    <nav>
      <a href="../index.html">首页</a>
      <a href="about.html">关于</a>
    </nav>
  </header>

  <main class="container">
    <section class="hero">
      <h2>关于本站</h2>
      <p>{{.Desc}}</p>
    </section>
  </main>

  <footer class="site-footer">
    <p>&copy; {{.Title}}. 保留所有权利。</p>
  </footer>

  <script src="../js/main.js"></script>
</body>
</html>
`

const headerHTML = `<!-- 复用页头片段：需要时复制进对应页面，或接入模板引擎统一注入 -->
<header class="site-header">
  <h1>{{.Title}}</h1>
  <nav>
    <a href="../index.html">首页</a>
    <a href="about.html">关于</a>
  </nav>
</header>
`

const footerHTML = `<!-- 复用页脚片段 -->
<footer class="site-footer">
  <p>&copy; {{.Title}}. 保留所有权利。</p>
</footer>
`

const mainCSS = `:root {
  --primary: #2563eb;
  --text: #1f2937;
  --muted: #6b7280;
  --bg: #f9fafb;
}

* { box-sizing: border-box; margin: 0; padding: 0; }

body {
  font-family: system-ui, -apple-system, "Segoe UI", sans-serif;
  color: var(--text);
  background: var(--bg);
  line-height: 1.6;
}

.site-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 2rem;
  background: #fff;
  border-bottom: 1px solid #e5e7eb;
}

.site-header nav a {
  margin-left: 1rem;
  color: var(--muted);
  text-decoration: none;
}

.site-header nav a:hover { color: var(--primary); }

.container { max-width: 960px; margin: 0 auto; padding: 2rem; }

.hero h2 { margin-bottom: 0.5rem; }
.hero p { color: var(--muted); }

.site-footer {
  padding: 1.5rem 2rem;
  text-align: center;
  color: var(--muted);
  border-top: 1px solid #e5e7eb;
}
`

const mainJS = `// 全站交互脚本入口
document.addEventListener("DOMContentLoaded", function () {
  console.log("页面已加载");
});
`

const assetsReadme = `# assets 目录

存放图片、字体、图标等静态资源，按子目录归类，例如：

- images/  图片
- fonts/   字体
- icons/   图标
`

const projectReadme = `# {{.Title}}

{{.Desc}}

## 目录结构（按功能拆分）

| 目录          | 职责                                   |
|---------------|----------------------------------------|
| index.html    | 首页入口                               |
| pages/        | 子页面                                 |
| components/   | 可复用的 HTML 片段（页头、页脚等）      |
| css/          | 样式表                                 |
| js/           | 交互脚本                               |
| assets/       | 图片、字体、图标等静态资源              |

## 本地预览

直接用浏览器打开 index.html，或在本目录运行：

` + "```" + `
npx serve .
` + "```" + `
`
