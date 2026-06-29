package dto

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=2,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Nickname string `json:"nickname" binding:"max=64"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty,len=11"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
	DeptID   uint   `json:"deptId" binding:"required"`
	RoleIds  []uint `json:"roleIds"`
	PostIds  []uint `json:"postIds"`
	Remark   string `json:"remark" binding:"max=500"`
}

type UpdateUserRequest struct {
	ID       uint   `json:"id" binding:"required"`
	Nickname string `json:"nickname" binding:"max=64"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty,len=11"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
	DeptID   uint   `json:"deptId"`
	RoleIds  []uint `json:"roleIds"`
	PostIds  []uint `json:"postIds"`
	Remark   string `json:"remark" binding:"max=500"`
}

type ResetPasswordRequest struct {
	ID       uint   `json:"id" binding:"required"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=128"`
}

type UpdateUserRolesRequest struct {
	ID      uint   `json:"id" binding:"required"`
	RoleIds []uint `json:"roleIds"`
}

type UpdateUserDeptRequest struct {
	ID     uint `json:"id" binding:"required"`
	DeptID uint `json:"deptId" binding:"required"`
}

type UserListRequest struct {
	Username string `json:"username" form:"username"`
	Phone    string `json:"phone" form:"phone"`
	Status   *int8  `json:"status" form:"status"`
	DeptID   uint   `json:"deptId" form:"deptId"`
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
}
