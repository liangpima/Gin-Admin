package dto

type CreateMemberRequest struct {
	Username string `json:"username" binding:"max=64"`
	Nickname string `json:"nickname" binding:"max=64"`
	Phone    string `json:"phone" binding:"required,len=11"`
	Avatar   string `json:"avatar" binding:"max=512"`
	Gender   int8   `json:"gender" binding:"oneof=0 1 2"`
	Birthday string `json:"birthday" binding:"omitempty"`
	LevelID  uint   `json:"levelId"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
	TagIds   []uint `json:"tagIds"`
	Remark   string `json:"remark" binding:"max=500"`
}

type UpdateMemberRequest struct {
	ID       uint   `json:"id" binding:"required"`
	Username string `json:"username" binding:"max=64"`
	Nickname string `json:"nickname" binding:"max=64"`
	Phone    string `json:"phone" binding:"omitempty,len=11"`
	Avatar   string `json:"avatar" binding:"max=512"`
	Gender   int8   `json:"gender" binding:"oneof=0 1 2"`
	Birthday string `json:"birthday" binding:"omitempty"`
	LevelID  uint   `json:"levelId"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
	TagIds   []uint `json:"tagIds"`
	Remark   string `json:"remark" binding:"max=500"`
}

type MemberListRequest struct {
	Phone    string `json:"phone" form:"phone"`
	Nickname string `json:"nickname" form:"nickname"`
	LevelID  uint   `json:"levelId" form:"levelId"`
	Status   *int8  `json:"status" form:"status"`
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
}

type UpdateMemberStatusRequest struct {
	ID     uint `json:"id" binding:"required"`
	Status int8 `json:"status" binding:"oneof=0 1"`
}

type UpdateMemberTagsRequest struct {
	ID     uint   `json:"id" binding:"required"`
	TagIds []uint `json:"tagIds"`
}

type CreateMemberLevelRequest struct {
	Name      string  `json:"name" binding:"required,max=64"`
	MinPoints int64   `json:"minPoints"`
	Discount  float64 `json:"discount" binding:"min=1,max=10"`
	Icon      string  `json:"icon" binding:"max=256"`
	Sort      int     `json:"sort"`
	Status    int8    `json:"status" binding:"oneof=0 1"`
}

type UpdateMemberLevelRequest struct {
	ID        uint    `json:"id" binding:"required"`
	Name      string  `json:"name" binding:"required,max=64"`
	MinPoints int64   `json:"minPoints"`
	Discount  float64 `json:"discount" binding:"min=1,max=10"`
	Icon      string  `json:"icon" binding:"max=256"`
	Sort      int     `json:"sort"`
	Status    int8    `json:"status" binding:"oneof=0 1"`
}

type MemberLevelListRequest struct {
	Name     string `json:"name" form:"name"`
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
}

type CreateMemberTagRequest struct {
	Name   string `json:"name" binding:"required,max=64"`
	Color  string `json:"color" binding:"max=20"`
	Sort   int    `json:"sort"`
	Status int8   `json:"status" binding:"oneof=0 1"`
}

type UpdateMemberTagRequest struct {
	ID     uint   `json:"id" binding:"required"`
	Name   string `json:"name" binding:"required,max=64"`
	Color  string `json:"color" binding:"max=20"`
	Sort   int    `json:"sort"`
	Status int8   `json:"status" binding:"oneof=0 1"`
}

type MemberTagListRequest struct {
	Name     string `json:"name" form:"name"`
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
}

type PointsLogListRequest struct {
	MemberID uint `json:"memberId" form:"memberId"`
	Type     int8 `json:"type" form:"type"`
	Page     int  `json:"page" form:"page"`
	PageSize int  `json:"pageSize" form:"pageSize"`
}
