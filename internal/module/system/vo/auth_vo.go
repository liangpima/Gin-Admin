package vo

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}

type UserInfoResponse struct {
	ID       uint         `json:"id"`
	Username string       `json:"username"`
	Nickname string       `json:"nickname"`
	Avatar   string       `json:"avatar"`
	Email    string       `json:"email"`
	Phone    string       `json:"phone"`
	Roles    []RoleInfo   `json:"roles"`
	Buttons  []string     `json:"buttons"`
	Menus    []MenuInfo   `json:"menus"`
}

type RoleInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type MenuInfo struct {
	ID       uint       `json:"id"`
	ParentID uint       `json:"parentId"`
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Component string   `json:"component"`
	Redirect string     `json:"redirect"`
	Icon     string     `json:"icon"`
	Title    string     `json:"title"`
	Type     int8       `json:"type"`
	Sort     int        `json:"sort"`
	IsCache  int8       `json:"isCache"`
	Visible  int8       `json:"visible"`
	Children []MenuInfo `json:"children"`
}
