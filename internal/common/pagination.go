package common

import "gorm.io/gorm"

type Pagination struct {
	Page     int
	PageSize int
	Offset   int
}

func NewPagination(page, pageSize int) *Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return &Pagination{
		Page:     page,
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
	}
}

func (p *Pagination) Apply(db *gorm.DB) *gorm.DB {
	return db.Offset(p.Offset).Limit(p.PageSize)
}
