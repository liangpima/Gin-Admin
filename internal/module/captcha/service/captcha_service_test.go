package service

import (
	"image"
	"image/color"
	"testing"
)

// TestRandomCharsNoDuplicate 验证生成的验证码字符互不重复。
//
// 该约束是功能正确性的前提：验证码为「按提示顺序依次点击」，
// 若出现重复字符（如 ABA），图上会有两个相同的 A，
// 用户无法分辨应先点击哪一个，只能靠猜，会直接导致验证失败。
func TestRandomCharsNoDuplicate(t *testing.T) {
	s := &captchaService{}
	const iterations = 2000

	for i := 0; i < iterations; i++ {
		chars, err := s.randomChars(charCount)
		if err != nil {
			t.Fatalf("第 %d 次生成失败: %v", i, err)
		}

		runes := []rune(chars)
		if len(runes) != charCount {
			t.Fatalf("期望 %d 个字符，实际 %d 个: %q", charCount, len(runes), chars)
		}

		seen := make(map[rune]bool, charCount)
		for _, ch := range runes {
			if seen[ch] {
				t.Fatalf("第 %d 次生成出现重复字符 %q: %q", i, ch, chars)
			}
			seen[ch] = true
		}
	}
}

// TestRandomCharsAllFromPool 验证生成的字符全部来自字符池
func TestRandomCharsAllFromPool(t *testing.T) {
	s := &captchaService{}

	pool := make(map[rune]bool, len(charPool))
	for _, ch := range charPool {
		pool[ch] = true
	}

	for i := 0; i < 500; i++ {
		chars, err := s.randomChars(charCount)
		if err != nil {
			t.Fatalf("生成失败: %v", err)
		}
		for _, ch := range chars {
			if !pool[ch] {
				t.Fatalf("字符 %q 不在字符池中: %q", ch, chars)
			}
		}
	}
}

// TestRandomCharsExceedsPool 验证请求长度超过字符池容量时返回错误，
// 而不是生成带重复字符的结果
func TestRandomCharsExceedsPool(t *testing.T) {
	s := &captchaService{}
	if _, err := s.randomChars(len(charPool) + 1); err == nil {
		t.Fatal("请求长度超过字符池容量时应返回错误")
	}
}

// newOpaqueCanvas 生成一张不透明画布，模拟真实背景（真实背景 alpha 恒为 255）
func newOpaqueCanvas() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, bgWidth, bgHeight))
	for y := 0; y < bgHeight; y++ {
		for x := 0; x < bgWidth; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	return img
}

// inkBounds 返回相对背景色之外的像素包围盒
func inkBounds(img *image.RGBA) (minX, minY, maxX, maxY int, ok bool) {
	minX, minY = bgWidth, bgHeight
	maxX, maxY = -1, -1
	for y := 0; y < bgHeight; y++ {
		for x := 0; x < bgWidth; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r == 0xffff && g == 0xffff && b == 0xffff {
				continue // 背景像素
			}
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	return minX, minY, maxX, maxY, maxX >= 0
}

// TestDrawCharCenteredOnPoint 验证字符墨迹的视觉中心与记录的目标点一致。
//
// Verify 是按 tolerancePx 比对「用户点击坐标」与「记录坐标」的。
// 如果字符画偏了，用户照着屏幕点击字符中心也会产生固定偏差，
// 偏差再叠加点击误差就会超出容差，导致明明点对了却验证失败。
func TestDrawCharCenteredOnPoint(t *testing.T) {
	s := &captchaService{}
	const px, py = 300, 100
	const maxOffset = 2 // 允许的居中误差（像素）

	for _, ch := range charPool {
		img := newOpaqueCanvas()
		s.drawChar(img, px, py, ch)

		minX, minY, maxX, maxY, ok := inkBounds(img)
		if !ok {
			t.Fatalf("字符 %q 未绘制出任何像素", ch)
		}

		centerX := float64(minX+maxX+1) / 2
		centerY := float64(minY+maxY+1) / 2
		dx := centerX - px
		dy := centerY - py
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}

		t.Logf("字符 %q 墨迹范围 x[%d,%d] y[%d,%d]，中心 (%.1f, %.1f)，相对目标点偏移 (%.1f, %.1f)",
			ch, minX, maxX, minY, maxY, centerX, centerY, dx, dy)

		if dx > maxOffset || dy > maxOffset {
			t.Errorf("字符 %q 墨迹中心偏移 (%.1f, %.1f) 超过允许值 %d，会挤占 Verify 的容差 %d",
				ch, dx, dy, maxOffset, tolerancePx)
		}
	}
}

// TestDrawnGlyphsMatchRecordedPoints 端到端几何校验：
// 完整走一遍「生成字符 -> 生成坐标 -> 渲染背景 -> 从图中还原字符位置」，
// 模拟用户照着自己看到的字符点击，检测到的视觉中心必须落在 Verify 的容差内。
func TestDrawnGlyphsMatchRecordedPoints(t *testing.T) {
	s := &captchaService{}
	const rounds = 20

	for round := 0; round < rounds; round++ {
		chars, err := s.randomChars(charCount)
		if err != nil {
			t.Fatalf("生成字符失败: %v", err)
		}
		points, err := s.randomPoints(charCount)
		if err != nil {
			t.Fatalf("生成坐标失败: %v", err)
		}

		img := s.generateBackground(chars, points)

		// 字符颜色固定为 RGB(10,40,100)，背景渐变/干扰线/噪点颜色均与之不同，
		// 因此可据此精确分离出字符墨迹
		type pixel struct{ x, y int }
		ink := make([]pixel, 0, 4096)
		for y := 0; y < bgHeight; y++ {
			for x := 0; x < bgWidth; x++ {
				r, g, b, _ := img.At(x, y).RGBA()
				if r == 10*257 && g == 40*257 && b == 100*257 {
					ink = append(ink, pixel{x, y})
				}
			}
		}
		if len(ink) == 0 {
			t.Fatalf("第 %d 轮：背景图中未找到任何字符墨迹", round)
		}

		// 按最近的目标点把墨迹归属到各字符，再求每个字符的视觉中心
		type acc struct{ sx, sy, n int }
		groups := make([]acc, len(points))
		for _, p := range ink {
			best, bestDist := -1, 0
			for i, pt := range points {
				dx, dy := p.x-pt.X, p.y-pt.Y
				d := dx*dx + dy*dy
				if best < 0 || d < bestDist {
					best, bestDist = i, d
				}
			}
			groups[best].sx += p.x
			groups[best].sy += p.y
			groups[best].n++
		}

		runes := []rune(chars)
		for i, g := range groups {
			if g.n == 0 {
				t.Fatalf("第 %d 轮：字符 %q 没有墨迹像素", round, string(runes[i]))
			}

			cx := float64(g.sx) / float64(g.n)
			cy := float64(g.sy) / float64(g.n)
			dx := cx - float64(points[i].X)
			dy := cy - float64(points[i].Y)
			if dx < 0 {
				dx = -dx
			}
			if dy < 0 {
				dy = -dy
			}

			if dx > tolerancePx || dy > tolerancePx {
				t.Fatalf("第 %d 轮字符 %q：视觉中心 (%.1f,%.1f) 距记录点 (%d,%d) 偏移 (%.1f,%.1f)，超出容差 %d",
					round, string(runes[i]), cx, cy, points[i].X, points[i].Y, dx, dy, tolerancePx)
			}
		}
	}
}
