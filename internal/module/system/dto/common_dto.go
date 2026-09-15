package dto

type IDRequest struct {
	ID uint `json:"id" binding:"required"`
}

type StatusRequest struct {
	ID uint `json:"id" binding:"required"`
	// 注意：不能用 required —— int8 的零值(0)会被 required 判为「未提供」，
	// 导致「停用」操作被参数校验直接拦下（返回 400）。
	Status int8 `json:"status" binding:"oneof=0 1"`
}

type DeleteRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}
