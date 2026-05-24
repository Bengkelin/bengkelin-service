package repository

import (
	"context"
	"math"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/pkg/helpers"
	"gorm.io/gorm"
)

// Common function to create in db
func Create(ctx context.Context, value interface{}) error {
	return db.GetDB().WithContext(ctx).Create(value).Error
}

// Common function to save in db
func Save(ctx context.Context, value interface{}) error {
	return db.GetDB().WithContext(ctx).Updates(value).Error
}

// Common function to get the first row
// Associations mean its relation to other
func First(ctx context.Context, where interface{}, out interface{}, associations []string) (notFound bool, err error) {
	db := db.GetDB().WithContext(ctx)
	for _, a := range associations {
		db = db.Preload(a)
	}
	err = db.Where(where).First(out).Error
	if err != nil {
		notFound = err == gorm.ErrRecordNotFound
	}
	return
}

func GetByID(ctx context.Context, where, out interface{}) (notFound bool, err error) {
	db := db.GetDB().WithContext(ctx)

	err = db.Where(where).Take(out).Error
	if err != nil {
		notFound = err == gorm.ErrRecordNotFound
	}
	return
}

// Common function to update in db
func Update(ctx context.Context, where, value interface{}) error {
	return db.GetDB().WithContext(ctx).Model(where).Updates(value).Error
}

// Common function to find in db
func Find(ctx context.Context, where interface{}, output interface{}, associations []string, orders ...string) error {
	db := db.GetDB().WithContext(ctx)
	for _, a := range associations {
		db = db.Preload(a)
	}
	db = db.Where(where)
	if len(orders) > 0 {
		for _, order := range orders {
			db = db.Order(order)
		}
	}
	return db.Find(output).Error
}

// Common function to paginate by model in db
func Query(ctx context.Context, where interface{}, output interface{}, pagination helpers.Pagination, associations []string) (*helpers.Pagination, error) {
	db := db.GetDB().WithContext(ctx)
	db.Scopes(paginate(where, &pagination, db))
	// preload the associations
	for _, a := range associations {
		db = db.Preload(a)
	}
	db.Find(output)
	pagination.Rows = output
	return &pagination, nil
}

// Func pagination for scope
func paginate(value interface{}, pagination *helpers.Pagination, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var totalRows int64
	db.Model(value).Count(&totalRows)

	pagination.TotalRows = totalRows
	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.GetLimit())))
	pagination.TotalPages = totalPages

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
	}
}

// Common function to delete by model in db
func DeleteByModel(ctx context.Context, model interface{}) (count int64, err error) {
	db := db.GetDB().WithContext(ctx).Delete(model)
	err = db.Error
	if err != nil {
		return
	}
	count = db.RowsAffected
	return
}

// Common function to delete by where in db
func DeleteByWhere(ctx context.Context, model, where interface{}) (count int64, err error) {
	db := db.GetDB().WithContext(ctx).Where(where).Delete(model)
	err = db.Error
	if err != nil {
		return
	}
	count = db.RowsAffected
	return
}

// Common function to delete by id in db
func DeleteByID(ctx context.Context, model interface{}, id uint64) (count int64, err error) {
	db := db.GetDB().WithContext(ctx).Where("id=?", id).Delete(model)
	err = db.Error
	if err != nil {
		return
	}
	count = db.RowsAffected
	return
}

// Common function to delete by ids (multiple) in db
func DeleteByIDS(ctx context.Context, model interface{}, ids []uint64) (count int64, err error) {
	db := db.GetDB().WithContext(ctx).Where("id in (?)", ids).Delete(model)
	err = db.Error
	if err != nil {
		return
	}
	count = db.RowsAffected
	return
}
