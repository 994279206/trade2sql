package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	// "github.com/flopp/go-findfont"
	// "github.com/goki/freetype/truetype"
	"github.com/trade2sql/internal/config"
	"github.com/trade2sql/internal/db"
	"github.com/trade2sql/internal/generator"
	"github.com/trade2sql/internal/resources"
)

// func init() {
// 	fontPath, err := findfont.Find("arial.ttf")
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("Found 'arial.ttf' in '%s'\n", fontPath)

// 	// load the font with the freetype library
// 	// 原作者使用的ioutil.ReadFile已经弃用
// 	fontData, err := os.ReadFile(fontPath)
// 	if err != nil {
// 		panic(err)
// 	}
// 	_, err = truetype.Parse(fontData)
// 	if err != nil {
// 		panic(err)
// 	}
// 	os.Setenv("FYNE_FONT", fontPath)
// 	os.Setenv("FYNE_FONT_MONOSPACE", fontPath)
// }

// StartGUI 启动GUI界面
func StartGUI() {
	a := app.New()
	a.SetIcon(resources.AppIcon)
	w := a.NewWindow("Trade2SQL - 数据库表结构生成工具")
	w.Resize(fyne.NewSize(1000, 700)) // 增加窗口大小以适应结构体预览

	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		cfg = config.Default()
	}

	// 数据库类型选择
	dbTypeSelect := widget.NewSelect([]string{"mysql", "postgres", "sqlite3"}, nil)
	dbTypeSelect.SetSelected(cfg.Database.Type)

	// 数据库连接字符串输入
	dbConnEntry := widget.NewEntry()
	dbConnEntry.SetText(cfg.Database.Connection)
	dbConnEntry.SetPlaceHolder("例如: user:password@tcp(localhost:3306)/database")

	// 表列表选择
	tableList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject { return widget.NewLabel("加载中...") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {},
	)

	// 选中的表名
	var selectedTable string

	// 表列表容器
	tableListContainer := container.NewVBox(
		widget.NewLabel("数据库表列表 (请先连接数据库)"),
		container.NewScroll(tableList),
	)

	// 输出包名
	packageNameEntry := widget.NewEntry()
	packageNameEntry.SetText(cfg.Generator.PackageName)
	packageNameEntry.SetPlaceHolder("输出的包名")

	// 标签格式
	tagFormatEntry := widget.NewEntry()
	tagFormatEntry.SetText(cfg.Generator.TagFormat)
	tagFormatEntry.SetPlaceHolder("标签格式，如: json,db")

	// 输出路径
	outputPathEntry := widget.NewEntry()
	str, _ := os.Getwd()
	outputPathEntry.SetText(str)
	outputPathEntry.SetPlaceHolder("输出文件路径")

	// 连接并获取表列表按钮
	connectBtn := widget.NewButton("连接数据库并获取表列表", func() {
		dbType := dbTypeSelect.Selected
		dbConn := dbConnEntry.Text

		database, err := db.Connect(dbType, dbConn)
		if err != nil {
			dialog.ShowError(fmt.Errorf("连接失败: %v", err), w)
			return
		}
		defer database.Close()

		// 获取表列表
		tables, err := database.GetTableList()
		if err != nil {
			dialog.ShowError(fmt.Errorf("获取表列表失败: %v", err), w)
			return
		}

		// 排序表名
		sort.Strings(tables)

		// 更新表列表
		tableList.Length = func() int { return len(tables) }
		tableList.CreateItem = func() fyne.CanvasObject { return widget.NewLabel("") }
		tableList.UpdateItem = func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(tables[id])
		}
		tableList.OnSelected = func(id widget.ListItemID) {
			selectedTable = tables[id]
			// if outputPathEntry.Text == "" {
			// 	outputPathEntry.SetText(fmt.Sprintf("%s_model.go", selectedTable))
			// 	// outputPathEntry.Disable()
			// }
		}
		tableList.Refresh()
		// dialog.ShowInformation("连接成功", fmt.Sprintf("成功获取到 %d 个表", len(tables)), w)
	})

	// 结构体预览区域
	structPreview := widget.NewMultiLineEntry()
	structPreview.SetPlaceHolder("生成的结构体将显示在这里")
	// structPreview.Disable() // 设置为只读模式

	// 预览区域容器
	// previewContainer := container.NewVBox(
	// 	widget.NewLabel("结构体预览"),
	// 	container.NewVScroll(structPreview),
	// )

	// 生成按钮
	generateBtn := widget.NewButton("生成结构体", func() {
		dbType := dbTypeSelect.Selected
		dbConn := dbConnEntry.Text
		packageName := packageNameEntry.Text
		tagFormat := tagFormatEntry.Text
		outputPath := outputPathEntry.Text

		if selectedTable == "" {
			dialog.ShowError(fmt.Errorf("请先选择一个表"), w)
			return
		}

		// // 如果输出路径为空，使用默认路径
		// if outputPath == "" {
		// 	outputPath = fmt.Sprintf("%s_model.go", selectedTable)
		// }

		// 确保输出路径是绝对路径
		var outPath string
		if !filepath.IsAbs(outputPath) {
			outPath, _ = filepath.Abs(outputPath)
			os.MkdirAll(outPath, os.ModePerm)
			outputPath = filepath.Join(outPath, fmt.Sprintf("%s_model.go", selectedTable))
		} else {
			outPath = outputPath
			os.MkdirAll(outPath, os.ModePerm)
			outputPath = filepath.Join(outPath, fmt.Sprintf("%s_model.go", selectedTable))
		}

		// 连接数据库
		database, err := db.Connect(dbType, dbConn)
		if err != nil {
			dialog.ShowError(fmt.Errorf("连接数据库失败: %v", err), w)
			return
		}
		defer database.Close()

		// 生成结构体
		genCfg := config.GeneratorConfig{
			PackageName: packageName,
			TagFormat:   tagFormat,
		}

		// 获取生成的结构体内容
		structContent, err := generator.GenerateStructContent(database, selectedTable, genCfg)
		if err != nil {
			dialog.ShowError(fmt.Errorf("生成结构体失败: %v", err), w)
			return
		}

		// 显示在预览区域
		structPreview.SetText(structContent)

		// 保存到文件
		err = os.WriteFile(outputPath, []byte(structContent), 0644)
		if err != nil {
			dialog.ShowError(fmt.Errorf("保存文件失败: %v", err), w)
			return
		}

		// 保存配置
		cfg.Database.Type = dbType
		cfg.Database.Connection = dbConn
		cfg.Generator.PackageName = packageName
		cfg.Generator.TagFormat = tagFormat
		err = config.Save("config.yaml", cfg)
		if err != nil {
			dialog.ShowError(fmt.Errorf("保存配置失败: %v", err), w)
		}

		dialog.ShowInformation("生成成功", fmt.Sprintf("已成功生成结构体到目录%v", outPath), w)
	})

	// 左侧面板 - 数据库连接和表列表
	leftPanel := container.NewVBox(
		widget.NewLabel("数据库连接"),
		widget.NewForm(
			widget.NewFormItem("数据库类型", dbTypeSelect),
			widget.NewFormItem("连接字符串", dbConnEntry),
		),
		connectBtn,
		widget.NewSeparator(),
		tableListContainer,
	)

	// 右侧面板 - 生成配置
	rightPanel := container.NewVBox(
		widget.NewLabel("生成配置"),
		widget.NewForm(
			widget.NewFormItem("包名", packageNameEntry),
			widget.NewFormItem("标签格式", tagFormatEntry),
			widget.NewFormItem("输出文件名", outputPathEntry),
		),
		generateBtn,
	)

	// 上部分 - 左右分栏
	topContent := container.NewHSplit(
		leftPanel,
		rightPanel,
	)
	topContent.SetOffset(0.4) // 左侧占40%

	// 主布局 - 上下分栏
	mainContent := container.NewVSplit(
		topContent,
		structPreview,
	)
	mainContent.SetOffset(0.5) // 上部分占50%

	w.SetContent(mainContent)
	w.ShowAndRun()
}
