package models

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"gin-blog/pkg/setting"
	"time"
)

var db *gorm.DB

type Model struct {
	ID         int `gorm:"primary_key" json:"id"`
	CreatedOn  int `json:"created_on"`
	ModifiedOn int `json:"modified_on"`
	DeletedOn int `json:"deleted_on"`
}

func init() {

	var (
		err                                               error
		dbType, dbName, user, password, host, tablePrefix string
	)

	sec, err := setting.Cfg.GetSection("database")
	if err != nil {
		log.Fatal(2, "Fail to get section 'database': %v", err)
	}

	dbType = sec.Key("TYPE").String()
	dbName = sec.Key("NAME").String()
	user = sec.Key("USER").String()
	password = sec.Key("PASSWORD").String()
	host = sec.Key("HOST").String()
	tablePrefix = sec.Key("TABLE_PREFIX").String()

	db, err = gorm.Open(dbType, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		user,
		password,
		host,
		dbName))
	if err != nil {
		log.Println(err)
	}

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return tablePrefix + defaultTableName
	}

	db.SingularTable(true)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

	db.Callback().Create().Replace("gorm:update_time_stamp", updateTimeStampForCreateCallback)
	db.Callback().Update().Replace("gorm:update_time_stamp", updateTimeStampForUpdateCallback)
	db.Callback().Delete().Replace("gorm:delete", deleteCallback)
}

func CloseDB() {
	defer db.Close()
}

// 在添加数据时, 自动设置 CreatedOn 和 ModifiedOn 为当前时间
func updateTimeStampForCreateCallback(scope *gorm.Scope) {
	// 检查是否有含有错误（db.Error）
	if !scope.HasError() {
		nowTime := time.Now().Unix()
		// scope.FieldByName 通过 scope.Fields() 获取所有字段，判断当前是否包含所需字段
		if createTimeField, ok := scope.FieldByName("CreatedOn"); ok {
			// field.IsBlank 可判断该字段的值是否为空
			if createTimeField.IsBlank {
				// field.Set 用于给该字段设置值，参数为 interface{}
				createTimeField.Set(nowTime)
			}
		}

		if modifyTimeField, ok := scope.FieldByName("ModifiedOn"); ok {
			if modifyTimeField.IsBlank {
				modifyTimeField.Set(nowTime)
			}
		}
	}
}

// 在修改数据时, 自动设置 ModifiedOn 为当前时间
func updateTimeStampForUpdateCallback(scope *gorm.Scope) {
	// scope.Get(...) 根据入参获取设置了字面值的参数，
	// 例如 gorm:update_column ，它会去查找含这个字面值的字段属性
	// scope.SetColumn(...) 假设没有指定 update_column 的字段，我们默认在更新回调设置 ModifiedOn 的值
	if _, ok := scope.Get("gorm:update_column"); !ok {
		scope.SetColumn("ModifiedOn", time.Now().Unix())
	}
}

// 软删除
func deleteCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		var extraOption string
		// scope.Get("gorm:delete_option") 检查是否手动指定了delete_option
		if str, ok := scope.Get("gorm:delete_option"); ok {
			extraOption = fmt.Sprint(str)
		}

		// scope.FieldByName("DeletedOn") 获取我们约定的删除字段，
		// 若存在则 UPDATE 软删除，若不存在则 DELETE 硬删除
		deleteOnField, hasDeletedOnField := scope.FieldByName("DeletedOn")

		if !scope.Search.Unscoped && hasDeletedOnField {
			scope.Raw(fmt.Sprintf(
				"UPDATE %v SET %v=%v%v%v",
				// scope.QuotedTableName() 返回引用的表名，这个方法 GORM 会根据自身逻辑对表名进行一些处理
				scope.QuotedTableName(),
				scope.Quote(deleteOnField.DBName),
				// scope.AddToVars 该方法可以添加值作为SQL的参数，也可用于防范SQL注入
				scope.AddToVars(time.Now().Unix()),
				// scope.CombinedConditionSql() 返回组合好的条件SQL
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		} else {
			scope.Raw(fmt.Sprintf(
				"DELETE FROM %v%v%v",
				scope.QuotedTableName(),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		}
	}
}

func addExtraSpaceIfExist(str string) string {
	if str != "" {
		return " " + str
	}
	return ""
}