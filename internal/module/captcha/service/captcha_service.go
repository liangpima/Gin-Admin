package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/big"
	"time"

	"go-admin/internal/cache"
	"go-admin/internal/module/captcha/model"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	captchaPrefix = "captcha:"
	captchaExpiry = 5 * time.Minute
	bgWidth       = 640
	bgHeight      = 200
	charCount     = 3
	charSize      = 90
	tolerancePx   = 30
)

var charPool = []rune("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")

type CaptchaService interface {
	Generate() (*model.CaptchaGenerateResponse, error)
	Verify(token string, points []model.Point) (*model.CaptchaVerifyResponse, error)
}

type captchaService struct{}

type captchaData struct {
	Points []model.Point `json:"points"`
	Chars  string        `json:"chars"`
}

func NewCaptchaService() CaptchaService {
	return &captchaService{}
}

func (s *captchaService) Generate() (*model.CaptchaGenerateResponse, error) {
	chars, err := s.randomChars(charCount)
	if err != nil {
		return nil, err
	}

	points, err := s.randomPoints(charCount)
	if err != nil {
		return nil, err
	}

	bgImg := s.generateBackground(chars, points)
	bgBase64 := imageToBase64(bgImg)

	token := generateToken()

	data := captchaData{Points: points, Chars: chars}
	dataBytes, _ := json.Marshal(data)
	if err := cache.Set(context.Background(), captchaPrefix+token, string(dataBytes), captchaExpiry); err != nil {
		return nil, fmt.Errorf("缓存验证码失败: %w", err)
	}

	return &model.CaptchaGenerateResponse{
		Token:    token,
		Bg:       "data:image/png;base64," + bgBase64,
		BgWidth:  bgWidth,
		BgHeight: bgHeight,
		Chars:    chars,
	}, nil
}

func (s *captchaService) Verify(token string, points []model.Point) (*model.CaptchaVerifyResponse, error) {
	key := captchaPrefix + token
	val, err := cache.Get(context.Background(), key)
	if err != nil {
		return &model.CaptchaVerifyResponse{
			Success: false,
			Message: "验证码已过期，请重新获取",
		}, nil
	}

	cache.Del(context.Background(), key)

	var data captchaData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return &model.CaptchaVerifyResponse{
			Success: false,
			Message: "验证码数据异常",
		}, nil
	}

	if len(points) != len(data.Points) {
		return &model.CaptchaVerifyResponse{
			Success: false,
			Message: "点击数量不正确",
		}, nil
	}

	for i, p := range points {
		expected := data.Points[i]
		dx := p.X - expected.X
		dy := p.Y - expected.Y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if dx > tolerancePx || dy > tolerancePx {
			return &model.CaptchaVerifyResponse{
				Success: false,
				Message: "验证失败，请重试",
			}, nil
		}
	}

	newToken := generateToken()

	// 记录「该 token 已通过人机校验」。
	//
	// 校验结果必须落盘，否则前端拿到的 newToken 只是一个无意义的随机串 ——
	// 登录接口无从判断它是否真的通过过验证，攻击者直接 POST /auth/login
	// 就能完全绕过验证码，人机校验形同虚设。
	if err := cache.Set(context.Background(), verifiedKey(newToken), "1", captchaExpiry); err != nil {
		return &model.CaptchaVerifyResponse{
			Success: false,
			Message: "验证状态保存失败，请重试",
		}, nil
	}

	return &model.CaptchaVerifyResponse{
		Success: true,
		Token:   newToken,
		Message: "验证成功",
	}, nil
}

func verifiedKey(token string) string {
	return captchaPrefix + "verified:" + token
}

// ConsumeVerifiedToken 消费一次性的人机校验凭证。
//
// 登录接口调用：凭证存在则删除并返回 true（一次性，防止重放），
// 不存在说明未通过验证或已用过，返回 false。
func ConsumeVerifiedToken(token string) bool {
	if token == "" {
		return false
	}
	ctx := context.Background()
	key := verifiedKey(token)

	exists, err := cache.Exists(ctx, key)
	if err != nil || !exists {
		return false
	}
	_ = cache.Del(ctx, key)
	return true
}

// randomChars 从字符池中不重复地随机抽取 n 个字符。
//
// 必须保证互不相同：本验证码是「按提示顺序依次点击」，
// 一旦出现重复字符（如 ABA），图上会存在两个相同的 A，
// 用户无法分辨应先点击哪一个，只能靠猜，会直接导致验证失败。
func (s *captchaService) randomChars(n int) (string, error) {
	if n > len(charPool) {
		return "", fmt.Errorf("验证码长度 %d 超过字符池容量 %d", n, len(charPool))
	}

	// 复制字符池后做 Fisher-Yates 洗牌，取前 n 个即为不放回抽样结果
	pool := make([]rune, len(charPool))
	copy(pool, charPool)

	for i := len(pool) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		k := j.Int64()
		pool[i], pool[k] = pool[k], pool[i]
	}

	return string(pool[:n]), nil
}

func (s *captchaService) randomPoints(n int) ([]model.Point, error) {
	points := make([]model.Point, n)
	marginX := 60
	marginY := 50
	for i := 0; i < n; i++ {
		x, err := randInt(marginX, bgWidth-marginX)
		if err != nil {
			return nil, err
		}
		y, err := randInt(marginY, bgHeight-marginY)
		if err != nil {
			return nil, err
		}
		for j := 0; j < i; j++ {
			dx := x - points[j].X
			dy := y - points[j].Y
			if dx < 0 {
				dx = -dx
			}
			if dy < 0 {
				dy = -dy
			}
			if dx < charSize && dy < charSize {
				x, _ = randInt(marginX, bgWidth-marginX)
				y, _ = randInt(marginY, bgHeight-marginY)
				j = -1
			}
		}
		points[i] = model.Point{X: x, Y: y}
	}
	return points, nil
}

func (s *captchaService) generateBackground(chars string, points []model.Point) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, bgWidth, bgHeight))

	for y := 0; y < bgHeight; y++ {
		for x := 0; x < bgWidth; x++ {
			r := uint8(220 + (x*7+y*3)%36)
			g := uint8(220 + (x*3+y*7)%36)
			b := uint8(220 + (x*5+y*5)%36)
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	for i := 0; i < 20; i++ {
		sx, _ := randInt(0, bgWidth)
		sy, _ := randInt(0, bgHeight)
		ex, _ := randInt(0, bgWidth)
		ey, _ := randInt(0, bgHeight)
		cr := uint8(150 + i*4)
		cg := uint8(150 + i*3)
		cb := uint8(150 + i*2)
		s.drawLine(img, sx, sy, ex, ey, color.RGBA{R: cr, G: cg, B: cb, A: 120})
	}

	for i := 0; i < 50; i++ {
		px, _ := randInt(0, bgWidth)
		py, _ := randInt(0, bgHeight)
		img.SetRGBA(px, py, color.RGBA{
			R: uint8(100 + i*2),
			G: uint8(100 + i*2),
			B: uint8(100 + i*2),
			A: 180,
		})
	}

	for i, ch := range chars {
		if i < len(points) {
			s.drawChar(img, points[i].X, points[i].Y, ch)
		}
	}

	return img
}

func (s *captchaService) drawChar(img *image.RGBA, x, y int, ch rune) {
	scale := 4
	face := basicfont.Face7x13

	// 先把字符绘制到临时画布上，便于测量其真实墨迹范围
	charImg := image.NewRGBA(image.Rect(0, 0, 12, 18))
	charDrawer := &font.Drawer{
		Dst:  charImg,
		Src:  image.NewUniform(color.RGBA{R: 255, G: 255, B: 255, A: 255}),
		Face: face,
		Dot:  fixed.P(2, 13),
	}
	charDrawer.DrawString(string(ch))

	// 求墨迹包围盒。字形并未填满 12x18 画布，且不同字符范围不同，
	// 因此必须按实际墨迹居中，否则字符会整体偏离目标点。
	minX, minY, maxX, maxY := 12, 18, -1, -1
	for dy := 0; dy < 18; dy++ {
		for dx := 0; dx < 12; dx++ {
			if _, _, _, a := charImg.At(dx, dy).RGBA(); a != 0 {
				if dx < minX {
					minX = dx
				}
				if dx > maxX {
					maxX = dx
				}
				if dy < minY {
					minY = dy
				}
				if dy > maxY {
					maxY = dy
				}
			}
		}
	}
	if maxX < 0 {
		return // 空白字形，理论上不会出现
	}

	// 让墨迹中心正好落在目标点上，保证用户点击字符中心即可命中 Verify 的容差
	offsetX := x - (minX+maxX+1)*scale/2
	offsetY := y - (minY+maxY+1)*scale/2

	for dy := minY; dy <= maxY; dy++ {
		for dx := minX; dx <= maxX; dx++ {
			if _, _, _, a := charImg.At(dx, dy).RGBA(); a == 0 {
				continue
			}
			for sy := 0; sy < scale; sy++ {
				for sx := 0; sx < scale; sx++ {
					px := offsetX + dx*scale + sx
					py := offsetY + dy*scale + sy
					if px >= 0 && px < bgWidth && py >= 0 && py < bgHeight {
						img.SetRGBA(px, py, color.RGBA{R: 10, G: 40, B: 100, A: 255})
					}
				}
			}
		}
	}
}

func (s *captchaService) drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx - dy

	for {
		if x0 >= 0 && x0 < bgWidth && y0 >= 0 && y0 < bgHeight {
			img.SetRGBA(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func imageToBase64(img image.Image) string {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func randInt(min, max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + min, nil
}
