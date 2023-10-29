package initialize

import (
	"encoding/json"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/spf13/viper"
	"mxshop_srvs/user_srv/global"
)

func GetEnvInfo(env string) bool {
	viper.AutomaticEnv()
	return viper.GetBool(env)
}
func InitConfig() {
	// 从配置文件中读取配置
	debug := GetEnvInfo("MXSHOP")
	configPath := "config-dev.yaml"
	if debug {
		configPath = "user_srv/config-dev.yaml"
	} else {
		configPath = "user_srv/config-prd.yaml"
	}
	v := viper.New()
	// 设置路径
	v.SetConfigFile(configPath)
	err := v.ReadInConfig()
	if err != nil {
		panic(err)
	}
	if err := v.Unmarshal(&global.NacosConfig); err != nil {
		panic(err)
	}

	// 通过nacos配置信息读取其他配置信息，db等等
	sc := []constant.ServerConfig{
		{
			IpAddr: global.NacosConfig.Host,
			Port:   global.NacosConfig.Port,
		},
	}

	cc := constant.ClientConfig{
		NamespaceId:         global.NacosConfig.NameSpace, //we can create multiple clients with different namespaceId to support multiple namespace.When namespace is public, fill in the blank string here.
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "tmp/nacos/log",
		CacheDir:            "tmp/nacos/cache",
		LogLevel:            "debug",
	}

	// Create config client for dynamic configuration
	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": sc,
		"clientConfig":  cc,
	})
	if err != nil {
		panic(err)
	}
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: global.NacosConfig.DataId,
		Group:  global.NacosConfig.Group,
	})

	if err != nil {
		panic(err)
	}
	//serverConfig := config.ServerConfig{}
	if err := json.Unmarshal([]byte(content), &global.ServerConfig); err != nil {
		panic(err)
	}

	fmt.Println("globalServer", global.ServerConfig)

}
