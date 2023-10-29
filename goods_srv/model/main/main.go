package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
)

func genMd5(code string) string {
	ans := md5.New()
	_, _ = io.WriteString(ans, code)
	return hex.EncodeToString(ans.Sum(nil))
}
func main() {
	//dsn := "root:123456@tcp(127.0.0.1:3306)/mxshop_goods_srv?charset=utf8mb4&parseTime=True&loc=Local"
	//newLogger := logger.New(
	//
	//	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	//	logger.Config{
	//		SlowThreshold: time.Second, // Slow SQL threshold
	//		LogLevel:      logger.Info, // Log level
	//		Colorful:      true,        // Disable color
	//	},
	//)
	//var err error
	//DB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
	//	NamingStrategy: schema.NamingStrategy{
	//		SingularTable: true,
	//	},
	//	Logger: newLogger,
	//})
	//if err != nil {
	//	panic(err)
	//}
	//_ = DB.AutoMigrate(
	//	&model.Category{},
	//	&model.Brand{},
	//	&model.GoodsCategoryBrand{},
	//	&model.Banner{},
	//	&model.Goods{},
	//)
}
