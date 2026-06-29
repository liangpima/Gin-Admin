package dto

type IDRequest struct {
	ID uint `json:"id" binding:"required"`
}

type StatusRequest struct {
	ID     uint `json:"id" binding:"required"`
	Status int8 `json:"status" binding:"required,oneof=0 1"`
}

type DeleteRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}
