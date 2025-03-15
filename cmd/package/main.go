package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	"os/exec"

	"github.com/trade2sql/pkg/build"
)

func main() {
	// 命令行参数
	outputDir := flag.String("output", "dist", "输出目录")
	version := flag.String("version", "0.1.0", "版本号")
	platform := flag.String("platform", "all", "目标平台 (windows-amd64, linux-amd64, darwin-amd64, all)")
	flag.Parse()

	// 设置构建信息
	build.Info.Version = *version
	build.Info.Commit = getGitCommit()
	build.Info.BuildTime = time.Now().Format(time.RFC3339)

	// 确定目标平台
	var platforms  []string
	if *platform == "all" {
		platforms = []string{
			"windows-amd64",
			"windows-386",
			"linux-amd64",
			"linux-arm64",
			"darwin-amd64",
			"darwin-arm64",
		}
	} else if *platform == "" {
		platforms = []string{build.GetCurrentPlatform()}
	} else {
		platforms = []string{*platform}
	}

	// 为每个平台构建
	for _, plat := range platforms {
		fmt.Printf("正在为 %s 平台构建...\n", plat)
		platOutputDir := filepath.Join(*outputDir, plat)
		err := build.BuildForPlatform(platOutputDir, plat)
		if err != nil {
			log.Fatalf("构建失败: %v", err)
		}

		// 复制配置文件
		err = copyFile("config.yaml", filepath.Join(platOutputDir, "config.yaml"))
		if err != nil {
			log.Printf("复制配置文件失败: %v", err)
		}

		fmt.Printf("构建完成: %s\n", platOutputDir)
	}
}

// getGitCommit 获取Git提交哈希
func getGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD") // 需要在文件顶部添加 "os/exec" 导入
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	// 需要在文件顶部添加 "strings" 导入
	return string(output[:len(output)-1])
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}