package configs

import (
	"expense-log/internal/model"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func InitConfigs(configPtr string) *model.Config {

	viper.SetConfigFile(configPtr)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 1. 读取文件
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("读取配置文件 [%s] 失败: %w", configPtr, err))
	}

	fmt.Printf("成功加载配置文件: %s\n", viper.ConfigFileUsed())

	// 2. 映射到结构体
	var globalConfigs model.Config
	if err := viper.Unmarshal(&globalConfigs); err != nil {
		panic(fmt.Errorf("渲染配置失败: %w", err))
	}

	return &globalConfigs
}
