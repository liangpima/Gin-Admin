package dto

type CreateRoleRequest struct {
	Name      string `json:"name" binding:"required,min=2,max=64"`
	Code      string `json:"code" binding:"required,min=2,max=64"`
	Sort      int    `json:"sort"`
	Status    int8   `json:"status" binding:"oneof=0 1"`
	DataScope int8   `json:"dataScope" binding:"oneof=1 2 3 4 5"`
	MenuIds   []uint `json:"menuIds"`
	Remark    string `json:"remark" binding:"max=500"`
}

type UpdateRoleRequest struct {
	ID        uint   `json:"id" binding:"required"`
	Name      string `json:"name" binding:"min=2,max=64"`
	Code      string `json:"code" binding:"min=2,max=64"`
	Sort      int    `json:"sort"`
	Status    int8   `json:"status" binding:"oneof=0 1"`
	DataScope int8   `json:"dataScope" binding:"oneof=1 2 3 4 5"`
	MenuIds   []uint `json:"menuIds"`
	Remark    string `json:"remark" binding:"max=500"`
}

type RoleListRequest struct {
	Name   string `json:"name" form:"name"`
	Code   string `json:"code" form:"code"`
	Status *int8  `json:"status" form:"status"`
	Page   int    `json:"page" form:"page"`
	PageSize int  `json:"pageSize" form:"pageSize"`
}
