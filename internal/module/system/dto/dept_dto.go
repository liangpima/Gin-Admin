package dto

type CreateDeptRequest struct {
	ParentID uint   `json:"parentId"`
	Name     string `json:"name" binding:"required,max=64"`
	Sort     int    `json:"sort"`
	Leader   string `json:"leader" binding:"max=64"`
	Phone    string `json:"phone" binding:"max=16"`
	Email    string `json:"email" binding:"max=128"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}

type UpdateDeptRequest struct {
	ID       uint   `json:"id" binding:"required"`
	ParentID uint   `json:"parentId"`
	Name     string `json:"name" binding:"max=64"`
	Sort     int    `json:"sort"`
	Leader   string `json:"leader" binding:"max=64"`
	Phone    string `json:"phone" binding:"max=16"`
	Email    string `json:"email" binding:"max=128"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}
