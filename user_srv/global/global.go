package global

import (
	"gorm.io/gorm"
	"mxshop_srvs/user_srv/config"
)

var (
	DB           *gorm.DB
	ServerConfig config.ServerConfig
	NacosConfig  *config.NacosConfig = &config.NacosConfig{}
)

//func init() {
//	dsn := "remoteUser:123456@tcp(127.0.0.1:3306)/mxshop_user_srv?charset=utf8mb4&parseTime=True&loc=Local"
//	newLogger := logger.New(
//		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
//		logger.Config{
//			SlowThreshold: time.Second, // Slow SQL threshold
//			LogLevel:      logger.Info, // Log level
//			Colorful:      true,        // Disable color
//		},
//	)
//	var err error
//	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
//		NamingStrategy: schema.NamingStrategy{
//			SingularTable: true,
//		},
//		Logger: newLogger,
//	})
//	if err != nil {
//		panic(err)
//	}
//	hasUser := DB.Migrator().HasTable(&model.User{})
//	if !hasUser {
//		err := DB.AutoMigrate(&model.User{})
//		if err != nil {
//			return
//		}
//	}
//}
