package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	_ "embed"

	"github.com/trade2sql/internal/config"
	"github.com/trade2sql/internal/db"
	"github.com/trade2sql/internal/generator"
	"github.com/trade2sql/internal/gui"
)

//go:embed config.yaml
var configFile string

func main() {
	// 命令行参数
	// configFile := flag.String("config", "config.yaml", "配置文件路径")
	dbType := flag.String("db", "", "数据库类型 (mysql, postgres, sqlite)")
	dbConn := flag.String("conn", "", "数据库连接字符串")
	table := flag.String("table", "", "表名")
	output := flag.String("output", "models", "输出文件路径")
	guiMode := flag.Bool("gui", true, "启动GUI模式")
	flag.Parse()

	// GUI模式
	if *guiMode {
		gui.StartGUI()
		return
	}

	// 命令行模式
	// 加载配置
	cfg, err := config.Load(configFile)
	if err != nil {
		// 如果配置文件不存在，使用默认配置
		if os.IsNotExist(err) {
			cfg = config.Default()
		} else {
			log.Fatalf("加载配置文件失败: %v", err)
		}
	}

	// 命令行参数覆盖配置文件
	if *dbType != "" {
		cfg.Database.Type = *dbType
	}
	if *dbConn != "" {
		cfg.Database.Connection = *dbConn
	}

	// 连接数据库
	database, err := db.Connect(cfg.Database.Type, cfg.Database.Connection)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer database.Close()

	// 生成结构体
	if *table != "" {
		outputPath := *output
		if outputPath == "" {
			outputPath = fmt.Sprintf("%s_model.go", *table)
		}

		err = generator.GenerateStruct(database, *table, outputPath, cfg.Generator)
		if err != nil {
			log.Fatalf("生成结构体失败: %v", err)
		}
		fmt.Printf("已成功生成结构体到 %s\n", outputPath)
	} else {
		fmt.Println("请指定表名 (-table)")
	}
}
