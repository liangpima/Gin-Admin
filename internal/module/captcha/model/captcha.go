package model

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type CaptchaGenerateResponse struct {
	Token    string `json:"token"`
	Bg       string `json:"bg"`
	BgWidth  int    `json:"bgWidth"`
	BgHeight int    `json:"bgHeight"`
	Chars    string `json:"chars"`
}

type CaptchaVerifyRequest struct {
	Token  string  `json:"token" binding:"required"`
	Points []Point `json:"points" binding:"required"`
}

type CaptchaVerifyResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	Message string `json:"message"`
}
