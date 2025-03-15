package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// BuildInfo 构建信息
type BuildInfo struct {
	Version   string
	Commit    string
	BuildTime string
}

// Info 当前构建信息
var Info BuildInfo

// BuildForPlatform 为指定平台构建
func BuildForPlatform(outputDir, platform string) error {
	platformOs := ""
	arch := ""

	switch platform {
	case "windows-amd64":
		platformOs = "windows"
		arch = "amd64"
	case "windows-386":
		platformOs = "windows"
		arch = "386"
	case "linux-amd64":
		platformOs = "linux"
		arch = "amd64"
	case "linux-arm64":
		platformOs = "linux"
		arch = "arm64"
	case "darwin-amd64":
		platformOs = "darwin"
		arch = "amd64"
	case "darwin-arm64":
		platformOs = "darwin"
		arch = "arm64"
	default:
		return fmt.Errorf("不支持的平台: %s", platform)
	}

	// 创建输出目录
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return err
	}

	// 构建文件名
	outputName := "trade2sql"
	if platformOs == "windows" {
		outputName += ".exe"
	}
	outputPath := filepath.Join(outputDir, outputName)

	// 设置环境变量
	env := append(os.Environ(),
		fmt.Sprintf("GOOS=%s", platformOs),
		fmt.Sprintf("GOARCH=%s", arch),
		fmt.Sprintf("CGO_ENABLED=%s", "0"),
	)

	// 构建命令
	ldflags := fmt.Sprintf("-X github.com/trade2sql/pkg/build.Info.Version=%s -X github.com/trade2sql/pkg/build.Info.Commit=%s -X github.com/trade2sql/pkg/build.Info.BuildTime=%s",
		Info.Version, Info.Commit, Info.BuildTime)

	cmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", outputPath, "./cmd/trade2sql")
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// GetCurrentPlatform 获取当前平台
func GetCurrentPlatform() string {
	return fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
}