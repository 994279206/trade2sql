package resources

import (
	"fyne.io/fyne/v2"
)

//go :generate fyne bundle -o bundled.go ./assets
// AppIcon 应用图标
var AppIcon fyne.Resource

func init() {
	AppIcon, _ = fyne.LoadResourceFromPath("./assets/icon.png")
	// 这里会在构建时由资源打包工具自动填充
}
