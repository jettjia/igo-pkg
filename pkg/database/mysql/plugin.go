package mysql

import (
	"errors"
	"time"

	"github.com/gookit/color"
	"gorm.io/gorm"
)

const (
	callBackBeforeName = "core:before"
	callBackAfterName  = "core:after"
	startTime          = "_start_time"
)

type TracePlugin struct{}

func (op *TracePlugin) Name() string {
	return "tracePlugin"
}

func (op *TracePlugin) Initialize(db *gorm.DB) (err error) {
	_ = db.Callback().Create().Before("sql_gorm:before_create").Register(callBackBeforeName, before)
	_ = db.Callback().Query().Before("sql_gorm:query").Register(callBackBeforeName, before)
	_ = db.Callback().Delete().Before("sql_gorm:before_delete").Register(callBackBeforeName, before)
	_ = db.Callback().Update().Before("sql_gorm:setup_reflect_value").Register(callBackBeforeName, before)
	_ = db.Callback().Row().Before("sql_gorm:row").Register(callBackBeforeName, before)
	_ = db.Callback().Raw().Before("sql_gorm:raw").Register(callBackBeforeName, before)

	db.Callback().Query().Before("sql_gorm:query").Register("disable_raise_record_not_found", func(d *gorm.DB) {
		d.Statement.RaiseErrorOnNotFound = false
	})

	_ = db.Callback().Create().After("sql_gorm:after_create").Register(callBackAfterName, after)
	_ = db.Callback().Query().After("sql_gorm:after_query").Register(callBackAfterName, after)
	_ = db.Callback().Delete().After("sql_gorm:after_delete").Register(callBackAfterName, after)
	_ = db.Callback().Update().After("sql_gorm:after_update").Register(callBackAfterName, after)
	_ = db.Callback().Row().After("sql_gorm:row").Register(callBackAfterName, after)
	_ = db.Callback().Raw().After("sql_gorm:raw").Register(callBackAfterName, after)
	return
}

var _ gorm.Plugin = &TracePlugin{}

func before(db *gorm.DB) {
	db.InstanceSet(startTime, time.Now())
}

func after(db *gorm.DB) {
	err := db.Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}

	sql := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	if err != nil {
		color.Red.Println("sql.err:", sql)
		panic(err)
	}
}
