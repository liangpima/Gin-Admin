package dto

type CreateMenuRequest struct {
	ParentID   uint   `json:"parentId"`
	Name       string `json:"name" binding:"required,max=64"`
	Path       string `json:"path" binding:"max=200"`
	Component  string `json:"component" binding:"max=200"`
	Redirect   string `json:"redirect" binding:"max=200"`
	Icon       string `json:"icon" binding:"max=64"`
	Title      string `json:"title" binding:"max=64"`
	Type       int8   `json:"type" binding:"required,oneof=0 1 2"`
	Permission string `json:"permission" binding:"max=200"`
	Sort       int    `json:"sort"`
	Visible    int8   `json:"visible" binding:"oneof=0 1"`
	Status     int8   `json:"status" binding:"oneof=0 1"`
	IsExternal int8   `json:"isExternal" binding:"oneof=0 1"`
	IsCache    int8   `json:"isCache" binding:"oneof=0 1"`
}

type UpdateMenuRequest struct {
	ID         uint   `json:"id" binding:"required"`
	ParentID   uint   `json:"parentId"`
	Name       string `json:"name" binding:"max=64"`
	Path       string `json:"path" binding:"max=200"`
	Component  string `json:"component" binding:"max=200"`
	Redirect   string `json:"redirect" binding:"max=200"`
	Icon       string `json:"icon" binding:"max=64"`
	Title      string `json:"title" binding:"max=64"`
	Type       int8   `json:"type" binding:"oneof=0 1 2"`
	Permission string `json:"permission" binding:"max=200"`
	Sort       int    `json:"sort"`
	Visible    int8   `json:"visible" binding:"oneof=0 1"`
	Status     int8   `json:"status" binding:"oneof=0 1"`
	IsExternal int8   `json:"isExternal" binding:"oneof=0 1"`
	IsCache    int8   `json:"isCache" binding:"oneof=0 1"`
}
