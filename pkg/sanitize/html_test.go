package sanitize

import (
	"strings"
	"sync"
	"testing"
)

// TestRichTextRemovesScript 脚本标签必须被移除。
//
// 这是本包存在的首要理由：协议内容由富文本编辑器产出后原样入库，
// 若不在此处净化，任何渲染点（v-html / H5 / 导出页）都会执行它。
func TestRichTextRemovesScript(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		wantKeep string // 净化后仍应保留的正文；为空表示不检查
	}{
		{"普通 script", `<p>正文</p><script>alert(1)</script>`, "正文"},
		{"大小写混淆", `<p>正文</p><SCRIPT>alert(1)</SCRIPT>`, "正文"},
		{"外链 script", `<script type="text/javascript" src="//evil.example/x.js"></script>`, ""},
		{"嵌套在合法标签内", `<div><p><script>alert(1)</script>正文</p></div>`, "正文"},
		{"未闭合 script", `<p>正文<script>alert(1)`, "正文"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := RichText(c.input)
			lower := strings.ToLower(got)

			if strings.Contains(lower, "<script") {
				t.Errorf("script 标签未被移除: %s", got)
			}
			if strings.Contains(lower, "javascript:") {
				t.Errorf("输出了 javascript: 协议: %s", got)
			}
			if c.wantKeep != "" && !strings.Contains(got, c.wantKeep) {
				t.Errorf("净化不应丢掉正文 %q，实际输出: %s", c.wantKeep, got)
			}
		})
	}
}

// TestRichTextRemovesEventHandlers 内联事件属性必须被移除。
//
// <img src=x onerror=...> 是最经典的绕过手法：完全不需要 <script> 标签，
// 只要图片加载失败就会执行，因此「只过滤 script 标签」是不够的。
func TestRichTextRemovesEventHandlers(t *testing.T) {
	inputs := []string{
		`<img src="x" onerror="alert(1)">`,
		`<img src=x onerror=alert(1)>`,
		`<div onmouseover="alert(1)">正文</div>`,
		`<p onclick="alert(1)">正文</p>`,
		`<body onload="alert(1)">`,
		`<svg onload="alert(1)"></svg>`,
		`<a href="https://example.com" onfocus="alert(1)">x</a>`,
		`<input onfocus="alert(1)" autofocus>`,
		`<marquee onstart="alert(1)">x</marquee>`,
	}

	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			got := strings.ToLower(RichText(in))
			for _, bad := range []string{"onerror", "onmouseover", "onclick", "onload", "onfocus", "onstart"} {
				if strings.Contains(got, bad) {
					t.Errorf("事件属性 %s 未被移除，输出: %s", bad, got)
				}
			}
		})
	}
}

// TestRichTextRemovesDangerousProtocols 危险协议必须被阻断。
//
// <a href="javascript:..."> 是"点击即执行"的载体；
// data:text/html 则可在部分场景下直接渲染出新文档。
// 其中一条刻意使用 HTML 实体编码（java&#115;cript:）——
// 若净化器在实体解码**之前**做判断就会被绕过。
func TestRichTextRemovesDangerousProtocols(t *testing.T) {
	inputs := []string{
		`<a href="javascript:alert(1)">点击</a>`,
		`<a href="JaVaScRiPt:alert(1)">点击</a>`,
		`<a href="java&#115;cript:alert(1)">点击</a>`,
		`<a href="  javascript:alert(1)">点击</a>`,
		`<img src="javascript:alert(1)">`,
		`<a href="data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==">x</a>`,
		`<a href="vbscript:msgbox(1)">x</a>`,
	}

	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			got := strings.ToLower(RichText(in))
			if strings.Contains(got, "javascript:") {
				t.Errorf("javascript: 协议未被阻断，输出: %s", got)
			}
			if strings.Contains(got, "data:text/html") {
				t.Errorf("data:text/html 未被阻断，输出: %s", got)
			}
			if strings.Contains(got, "vbscript:") {
				t.Errorf("vbscript: 协议未被阻断，输出: %s", got)
			}
		})
	}
}

// TestRichTextKeepsSafeMarkup 正常富文本必须保留 —— 净化不能把内容吃掉。
// 覆盖协议类文档的常见元素：标题、加粗、列表、表格、链接、图片、引用、代码。
func TestRichTextKeepsSafeMarkup(t *testing.T) {
	input := `<h1>用户协议</h1>
<p>欢迎使用<strong>本服务</strong>。</p>
<ul><li>第一条</li><li>第二条</li></ul>
<ol><li>步骤一</li></ol>
<blockquote>提示</blockquote>
<table><thead><tr><th>项目</th></tr></thead><tbody><tr><td>值</td></tr></tbody></table>
<p><a href="https://example.com/doc">相关文档</a></p>
<p><img src="https://example.com/a.png" alt="示意图"></p>
<pre><code>示例代码</code></pre>`

	got := RichText(input)

	for _, want := range []string{
		"<h1>", "<strong>", "<ul>", "<li>", "<ol>", "<blockquote>",
		"<table>", "<th>", "<td>", "<pre>", "<code>",
		"用户协议", "本服务", "相关文档", "示意图", "示例代码",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("净化不应移除 %q，实际输出: %s", want, got)
		}
	}

	// 外链必须带 nofollow / noreferrer：
	//   nofollow   —— 防协议页被用来堆外链
	//   noreferrer —— 防 tabnabbing（目标页通过 window.opener 改写本页）
	if !strings.Contains(got, "nofollow") {
		t.Errorf("外链应带 rel=nofollow，实际输出: %s", got)
	}
	if !strings.Contains(got, "noreferrer") {
		t.Errorf("外链应带 rel=noreferrer，实际输出: %s", got)
	}
}

// TestRichTextRemovesEmbedding 嵌入与文档级标签必须移除。
//
// 协议页不需要 iframe/video 嵌入或整页样式覆盖，
// 留着它们等于给钓鱼、点击劫持、样式劫持留位置。
func TestRichTextRemovesEmbedding(t *testing.T) {
	inputs := []string{
		`<iframe src="https://evil.example"></iframe>`,
		`<object data="x.swf"></object>`,
		`<embed src="x.swf">`,
		`<form action="https://evil.example"><input name="p" type="password"><button>提交</button></form>`,
		`<style>body{display:none}</style>`,
		`<div style="position:fixed;top:0;left:0;width:100%;height:100%">遮罩</div>`,
		`<meta http-equiv="refresh" content="0;url=https://evil.example">`,
		`<link rel="stylesheet" href="https://evil.example/x.css">`,
		`<base href="https://evil.example/">`,
	}

	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			got := strings.ToLower(RichText(in))
			for _, bad := range []string{
				"<iframe", "<object", "<embed", "<form", "<style", "<meta", "<link", "<base", "<input", "<button",
			} {
				if strings.Contains(got, bad) {
					t.Errorf("危险标签 %s 未被移除，输出: %s", bad, got)
				}
			}
			// 内联 style 同样要去掉：可用于整页遮罩伪造界面
			if strings.Contains(got, "style=") {
				t.Errorf("内联 style 未被移除，输出: %s", got)
			}
		})
	}
}

// TestRichTextMalformedHTML 畸形 HTML 不得绕过净化，也不得 panic。
//
// 攻击者常用的手法就是构造解析器与过滤器理解不一致的输入
// （标签截断、属性引号不闭合、多余尖括号）。
func TestRichTextMalformedHTML(t *testing.T) {
	inputs := []string{
		`<p>未闭合<div><span>嵌套错乱`,
		`<p <script>alert(1)</script>>`,
		`<<script>alert(1)//<</script>`,
		`<img src="x" """><script>alert(1)</script>`,
		`<a href="https://a.com" "><script>alert(1)</script>`,
		`<div><p>正文</p><script>alert(1)</script></div></p></div>`,
		"<p>正文</p>\x00<script>alert(1)</script>",
	}

	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			got := RichText(in)
			if strings.Contains(strings.ToLower(got), "<script") {
				t.Errorf("畸形输入绕过了净化，输出: %s", got)
			}
			// 说明：输出中允许出现 alert(1) 这类**纯文本**残留，例如
			//   `<p <script>alert(1)</script>>`  →  `<p>alert(1)&gt;`
			// 这是安全的：标签已被移除、尖括号已转义（注意输出里的 `&gt;`），
			// 浏览器只会把它当普通文字显示。
			// 所以这里只断言「不存在标签形式」，而不是断言「字符串不含 alert」——
			// 后者会误伤本来就该保留的正常文本（协议里写 "alert(1)" 是合法的）。
		})
	}
}

// TestRichTextEmptyInput 空输入返回空串，不产生任何标签。
func TestRichTextEmptyInput(t *testing.T) {
	for _, in := range []string{"", "   ", "\n\t", "  \n  "} {
		if got := RichText(in); got != "" {
			t.Errorf("空白输入应返回空串，输入 %q got %q", in, got)
		}
	}
}

// TestRichTextPlainTextUnchanged 纯文本不受影响（净化不应改动无标签内容）。
func TestRichTextPlainTextUnchanged(t *testing.T) {
	in := "这是一段没有标签的协议文本。"
	if got := RichText(in); got != in {
		t.Errorf("纯文本不应被改动，got %q", got)
	}
}

// TestRichTextConcurrent 策略在进程内共享，必须能被并发调用。
//
// 服务端每个请求都会走到这里，Policy 只在首次调用时构建（sync.Once），
// 之后只读。本用例在 `go test -race` 下才有完整意义。
func TestRichTextConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = RichText(`<p>正文</p><script>alert(1)</script><img src=x onerror=alert(1)>`)
		}()
	}
	wg.Wait()
}
