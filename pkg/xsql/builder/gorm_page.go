package builder

import (
	"math"

	"gorm.io/gorm"
)

// GormPaginate sql_gorm page
func GormPaginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// CeilPageNum the number of pagination is calculated
func CeilPageNum(total int64, pageSize int) int64 {
	return int64(int(math.Ceil(float64(total) / float64(pageSize))))
}
