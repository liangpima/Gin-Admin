// Package sanitize 提供富文本内容的净化能力。
//
// 存在的理由：后台的部分内容字段（协议、公告等）由富文本编辑器产出，
// 本质是**原始 HTML**，会被原样入库、原样返回给调用方。
// 一旦任何前端或第三方端把它当 HTML 渲染（v-html、H5、打印页、小程序），
// 内容里注入的 <img onerror> / <a href="javascript:"> 就会执行。
//
// 本项目还有两个放大因素，使这件事不能只当作"低危的展示问题"：
//   - token 存在**非 httpOnly** 的 Cookie 里，JS 可直接读取 ——
//     一次 XSS 就等于账号被接管，而不是"页面被改两下"；
//   - 部分内容表是**全局表**（如 sys_agreement 无 tenant_id），
//     一个租户写入的恶意内容会影响所有租户。
//
// 净化放在**入库时**而不是渲染时，是刻意的选择：渲染点会随需求不断增加
// （管理页回显、H5、导出、第三方对接），只在入库处把关，
// 才能做到"不管将来加多少个渲染点都安全"。
package sanitize

import (
	"strings"
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

// richTextPolicy 富文本净化策略，进程内只构建一次。
//
// 基础策略直接用 bluemonday.UGCPolicy()：它是专为「用户生成内容」设计的
// 白名单策略，允许常见富文本标签（标题/段落/列表/表格/链接/图片/代码块），
// 同时移除 script、style、iframe、form、object 等危险标签
// 以及全部 on* 事件属性，并限制 URL 协议（javascript: 之类的会被丢弃）。
//
// 为什么不自研净化器：手写 HTML 白名单要处理标签闭合、属性转义、
// 畸形嵌套、解析器差异等大量绕过手法，是典型的"看起来能用、实则到处漏"。
// 安全组件应当用被广泛审计过的实现。
var (
	richTextOnce   sync.Once
	richTextPolicy *bluemonday.Policy
)

func policy() *bluemonday.Policy {
	richTextOnce.Do(func() {
		p := bluemonday.UGCPolicy()

		// 外链加固（两个都作用于 <a>，由 bluemonday 自动补 rel，而不是拒绝链接）：
		//   nofollow   —— 避免协议页被用来堆外链做 SEO 滥用
		//   noreferrer —— 避免 tabnabbing：目标页能通过 window.opener 改写本页
		p.RequireNoFollowOnLinks(true)
		p.RequireNoReferrerOnLinks(true)

		richTextPolicy = p
	})
	return richTextPolicy
}

// RichText 净化后台录入的富文本 HTML，返回可安全渲染的内容。
//
// 采用白名单：不在名单内的标签会被移除；具体某个标签的文本子节点是否保留
// 由 bluemonday 决定（不同标签行为不同），因此**不要依赖它做"提取纯文本"**，
// 该行为由 html_test.go 中的用例固定。
//
// 空白输入返回空串，不引入任何标签。
func RichText(input string) string {
	if strings.TrimSpace(input) == "" {
		return ""
	}
	return policy().Sanitize(input)
}
