package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config 应用配置
type Config struct {
	Database  DatabaseConfig  `yaml:"database"`
	Generator GeneratorConfig `yaml:"generator"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type       string `yaml:"type"`
	Connection string `yaml:"connection"`
}

// GeneratorConfig 生成器配置
type GeneratorConfig struct {
	PackageName string `yaml:"package_name"`
	TagFormat   string `yaml:"tag_format"`
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// Save 保存配置到文件
func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}

// Default 返回默认配置
func Default() *Config {
	return &Config{
		Database: DatabaseConfig{
			Type:       "mysql",
			Connection: "root:password@tcp(localhost:3306)/database",
		},
		Generator: GeneratorConfig{
			PackageName: "model",
			TagFormat:   "json,db",
		},
	}
}