/*
Copyright © 2023 Xu Wu <ixw1991@126.com>
Use of this source code is governed by a MIT style
license that can be found in the LICENSE file.
*/
package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"

	"github.com/imxw/gitlab-scaffold/internal/gitlabx"
	"github.com/imxw/gitlab-scaffold/internal/scaffold"
)

var globalConfig Config

type Config interface {
	GetGitlab() gitlabx.Config
	GetTemplate() scaffold.Config
}

type configImpl struct {
	Gitlab   gitlabx.Config  `mapstructure:"gitlab"`
	Template scaffold.Config `mapstructure:"template"`
}

func (c *configImpl) GetGitlab() gitlabx.Config {
	return c.Gitlab
}

func (c *configImpl) GetTemplate() scaffold.Config {
	return c.Template
}

// LoadConfig 加载配置文件并返回配置对象
func LoadConfig() (Config, error) {

	// 加载配置
	if err := initConfig(); err != nil {
		return nil, err
	}
	// 读取配置项
	config := &configImpl{}
	err := viper.Unmarshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	// 验证配置项是否存在
	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %v", err)
	}

	return config, nil
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	configFile := viper.GetString("config")
	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		// Search config in home directory with name ".glfast" (without extension).
		viper.AddConfigPath(filepath.Join(home, ".glfast"))
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("gl")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		viper.SetDefault("gitlab", map[string]string{"baseurl": "https://gitlab.com"})
		viper.SetDefault("template", map[string]interface{}{
			"extensions":        []string{},
			"base64_extensions": []string{},
			"files":             []string{},
		})
	}
	//  else {
	// 	fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	// }

	return nil

}

func InitializeConfig() error {
	if globalConfig != nil {
		log.Println("Config is already initialized, ignoring")
		return nil
	}

	cfg, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	globalConfig = cfg
	return nil
}

// C returns the global configuration object.
func C() Config {
	if globalConfig == nil {
		panic("Config not initialized")
	}
	return globalConfig
}
