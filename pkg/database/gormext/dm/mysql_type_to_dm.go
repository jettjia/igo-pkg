package dm

import (
	"errors"
	"fmt"
	"gorm.io/gorm/schema"
	"regexp"
	"strings"
)

func MysqlGormType2Dm(field *schema.Field) (returnField *schema.Field, err error) {
	var (
		sqlGORMDataType string
		sqlDataType     string
		sqlType         string
	)
	sqlGORMDataType = strings.ToLower(string(field.GORMDataType))
	sqlDataType = strings.ToLower(string(field.DataType))

	// mysql 类型转 dm类型
	// 如果是 int, uint;去掉指定的长度。比如：bigint(20) --> bigint; bigint(20) unsigned --> bigint unsigned
	switch sqlGORMDataType {
	case "int":
		// 移除长度
		sqlType = RemoveParentheses(sqlDataType)

		// mysql type --> dm type
		sqlType, err = MysqlType2Dm(sqlType)
		if err != nil {
			return nil, err
		}
	case "uint":
		// 移除长度
		sqlType = RemoveParentheses(sqlDataType)

		// mysql type --> dm type
		sqlType, err = MysqlType2Dm(sqlType)
		if err != nil {
			return nil, err
		}
	case "float":
		// 移除长度
		sqlType = RemoveParentheses(sqlDataType)

		// mysql type --> dm type
		sqlType, err = MysqlType2Dm(sqlType)
		if err != nil {
			return nil, err
		}
	default:
		// 如果是varchar, char 指定了长度，需要把长度跟着返回
		if strings.HasPrefix(sqlDataType, "varchar") || strings.HasPrefix(sqlDataType, "char") {
			return field, nil
		}

		// mysql type --> dm type
		sqlType, err = MysqlType2Dm(sqlDataType)
		if err != nil {
			return nil, err
		}
	}

	field.DataType = schema.DataType(sqlType)

	return field, nil
}

func MysqlType2Dm(mysqlType string) (dm string, err error) {
	if len(mysqlType) == 0 {
		return dm, errors.New(fmt.Sprintf("Func MysqlType2Dm， param[%s]: Request parameter cannot be empty", mysqlType))
	}

	mysqlType = strings.ToUpper(mysqlType)

	dmTypeList := mysqlAndDmTypeList()
	for k, v := range dmTypeList {
		if k == mysqlType {
			return v, nil
		}
	}

	return dm, errors.New(fmt.Sprintf("[%s] Type not found", mysqlType))
}

func mysqlAndDmTypeList() map[string]string {
	mysqlType := make(map[string]string)
	mysqlType["SET"] = "CHAR"
	mysqlType["POLYGON"] = "SYSGEO.ST POLYGON"
	mysqlType["INT UNSIGNED"] = "BIGINT"
	mysqlType["JSON"] = "VARCHAR"
	mysqlType["MEDIUMTEXT"] = "BIGINT"
	mysqlType["MULTIPOLYGON"] = "SYSGEO.ST MULTIPOLYGON"
	mysqlType["BIGINT UNSIGNED"] = "BIGINT"
	mysqlType["YEAR"] = "CHAR"
	mysqlType["TIMESTAMP"] = "TIMESTAMP"
	mysqlType["TINYINT UNSIGNED"] = "INT"
	mysqlType["SMALLINT UNSIGNED"] = "INT"
	mysqlType["DOUBLE"] = "NUMBER"
	mysqlType["TINYTEXT"] = "CLOB"
	mysqlType["LONGBLOB"] = "BLOB"
	mysqlType["GEOMETRY"] = "SYSGEO.ST GEOMETRY"
	mysqlType["GEOMETRYCOLLECTION"] = "SYSGEO.ST GEOMCOLLECTION"
	mysqlType["ENUM"] = "ENUM"
	mysqlType["LONGVARCHAR"] = "LONGVARCHAR"
	mysqlType["MULTIPOINT"] = "SYSGEO.ST MULTIPOINT"
	mysqlType["LONGTEXT"] = "CLOB"
	mysqlType["LINESTRING"] = "SYSGEO.ST LINESTRING"
	mysqlType["TIME"] = "TIME"
	mysqlType["MEDIUMINT"] = "INT"
	mysqlType["BIT"] = "BIT"
	mysqlType["MEDIUMINT UNSIGNED"] = "MEDIUMINT"
	mysqlType["BOOLEAN"] = "MEDIUMINT"
	mysqlType["MULTILINESTRING"] = "SYSGEO.ST MULTILINESTRING"
	mysqlType["DATETIME"] = "TIMESTAMP"
	mysqlType["MEDIUMBLOB"] = "BLOB"
	mysqlType["TINYBLOB"] = "BLOB"
	mysqlType["BOOL"] = "BIT"
	mysqlType["POINT"] = "SYSGEO.ST POINT"
	mysqlType["TINYINT"] = "TINYINT"
	mysqlType["BIGINT"] = "BIGINT"
	mysqlType["LONG VARBINARY"] = "BLOB"
	mysqlType["BLOB"] = "BLOB"
	mysqlType["VARBINARY"] = "VARBINARY"
	mysqlType["BINARY"] = "BINARY"
	mysqlType["LONG VARCHAR"] = "CLOB"
	mysqlType["TEXT"] = "TEXT"
	mysqlType["CHAR"] = "CHAR"
	mysqlType["NUMERIC"] = "NUMERIC"
	mysqlType["DECIMAL"] = "DECIMAL"
	mysqlType["INTEGER"] = "INTEGER"
	mysqlType["INT"] = "INT"
	mysqlType["SMALLINT"] = "SMALLINT"
	mysqlType["FLOAT"] = "REAL"
	mysqlType["DOUBLE PRECISION"] = "DOUBLE PRECISION"
	mysqlType["REAL"] = "DOUBLE PRECISION"
	mysqlType["VARCHAR"] = "VARCHAR"
	mysqlType["DATE"] = "DATE"

	return mysqlType
}

func RemoveParentheses(strInfo string) string {
	strReg := "\\(.*?\\)|\\{.*?}|\\[.*?]|（.*?）"
	reg := regexp.MustCompile(strReg)
	return reg.ReplaceAllString(strInfo, "")
}
