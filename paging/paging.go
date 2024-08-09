package paging

import (
	"gorm.io/gen/field"
)

type Paging interface {
	OrderBy() string
	SetTotal(int)
	GetTotal() int
	OffsetLimit() (offset, limit int)
}

func New() Paging {
	return nil
}

func FiltrWhere(field field.Expr, value any) field.Expr {
	return nil
}

func QueryPaging(query any, page Paging) error {
	_, _ = page.OffsetLimit()
	return nil
}
