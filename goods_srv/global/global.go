package global

import (
	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"
	"mxshop_srvs/goods_srv/config"
)

var (
	DB           *gorm.DB
	ServerConfig config.ServerConfig
	NacosConfig  *config.NacosConfig = &config.NacosConfig{}
	EsClient     *elastic.Client
)
