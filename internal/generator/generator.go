package generator

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
	"unicode"

	"github.com/trade2sql/internal/config"
	"github.com/trade2sql/internal/db"
)

// 结构体模板
const structTemplate = `// 代码由 trade2sql 自动生成
package {{.PackageName}}

{{if .Imports}}
import (
	{{range .Imports}}{{.}}
	{{end}}
)
{{end}}

// {{.StructName}} 对应数据库表 {{.TableName}}
type {{.StructName}} struct {
{{range .Fields}}	{{.Name}} {{.Type}} {{.Tag}} {{if .Comment}}// {{.Comment}}{{end}}
{{end}}
}
`

// TemplateData 模板数据
type TemplateData struct {
	PackageName string
	Imports     []string
	StructName  string
	TableName   string
	Fields      []FieldData
}

// FieldData 字段数据
type FieldData struct {
	Name    string
	Type    string
	Tag     string
	Comment string
}

// GenerateStructContent 生成结构体内容并返回字符串
func GenerateStructContent(database *db.Database, tableName string, cfg config.GeneratorConfig) (string, error) {
	columns, err := database.GetTableInfo(tableName)
	if err != nil {
		return "", err
	}

	// 准备模板数据
	data := TemplateData{
		PackageName: cfg.PackageName,
		StructName:  toUpperCamelCase(tableName),
		TableName:   tableName,
		Imports:     []string{},
	}

	// 处理字段
	for _, col := range columns {
		field := FieldData{
			Name:    toUpperCamelCase(col.Name),
			Type:    mapSQLTypeToGoType(col.Type, col.IsNullable),
			Comment: col.Comment,
		}

		// 添加必要的导入
		if strings.Contains(field.Type, "time.Time") && !containsString(data.Imports, `"time"`) {
			data.Imports = append(data.Imports, `"time"`)
		}

		// 生成标签
		tags := []string{}
		for _, tag := range strings.Split(cfg.TagFormat, ",") {
			tags = append(tags, fmt.Sprintf(`%v:"%s"`, tag, col.Name))
		}
		if col.IsPrimary {
			tags = append(tags, `primary:"true"`)
		}

		if len(tags) > 0 {
			field.Tag = fmt.Sprintf("`%s`", strings.Join(tags, " "))
		}

		data.Fields = append(data.Fields, field)
	}

	// 渲染模板
	tmpl, err := template.New("struct").Parse(structTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GenerateStruct 生成结构体并写入文件
func GenerateStruct(database *db.Database, tableName, outputPath string, cfg config.GeneratorConfig) error {
	content, err := GenerateStructContent(database, tableName, cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, []byte(content), 0644)
}

// toUpperCamelCase 转换为大驼峰命名
func toUpperCamelCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == ' '
	})

	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}

	return strings.Join(words, "")
}

// mapSQLTypeToGoType 将SQL类型映射到Go类型
func mapSQLTypeToGoType(sqlType string, isNullable bool) string {
	sqlType = strings.ToLower(sqlType)

	// 基本类型映射
	var goType string
	switch {
	case strings.Contains(sqlType, "int"):
		goType = "int64"
	case strings.Contains(sqlType, "float") || strings.Contains(sqlType, "double") || strings.Contains(sqlType, "decimal"):
		goType = "float64"
	case strings.Contains(sqlType, "bool"):
		goType = "bool"
	case strings.Contains(sqlType, "char") || strings.Contains(sqlType, "text") || strings.Contains(sqlType, "varchar"):
		goType = "string"
	case strings.Contains(sqlType, "date") || strings.Contains(sqlType, "time"):
		goType = "time.Time"
	case strings.Contains(sqlType, "blob") || strings.Contains(sqlType, "binary"):
		goType = "[]byte"
	default:
		goType = "interface{}"
	}

	// 处理可空类型
	if isNullable {
		switch goType {
		case "int64":
			return "*int64"
		case "float64":
			return "*float64"
		case "bool":
			return "*bool"
		case "time.Time":
			return "*time.Time"
		case "string":
			return "*string"
		default:
			return goType
		}
	}

	return goType
}

// containsString 检查字符串切片是否包含指定字符串
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
