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
	return &model.CaptchaVerifyResponse{
		Success: true,
		Token:   newToken,
		Message: "验证成功",
	}, nil
}

func (s *captchaService) randomChars(n int) (string, error) {
	poolSize := big.NewInt(int64(len(charPool)))
	result := make([]rune, n)
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, poolSize)
		if err != nil {
			return "", err
		}
		result[i] = charPool[idx.Int64()]
	}
	return string(result), nil
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
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{R: 10, G: 40, B: 100, A: 255}),
		Face: face,
		Dot:  fixed.P(0, 0),
	}

	charImg := image.NewRGBA(image.Rect(0, 0, 12, 18))
	charDrawer := &font.Drawer{
		Dst:  charImg,
		Src:  image.NewUniform(color.RGBA{R: 255, G: 255, B: 255, A: 255}),
		Face: face,
		Dot:  fixed.P(2, 13),
	}
	charDrawer.DrawString(string(ch))

	offsetX := x - 6*scale/2
	offsetY := y - 9*scale/2

	for dy := 0; dy < 18; dy++ {
		for dx := 0; dx < 12; dx++ {
			_, _, _, a := charImg.At(dx, dy).RGBA()
			if a == 0 {
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

	radius := charSize/2 + 4
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy > radius*radius {
				continue
			}
			px, py := x+dx, y+dy
			if px >= 0 && px < bgWidth && py >= 0 && py < bgHeight {
				_, _, _, a := img.At(px, py).RGBA()
				if a < 128 {
					img.SetRGBA(px, py, color.RGBA{R: 255, G: 255, B: 255, A: 180})
				}
			}
		}
	}

	_ = d
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
