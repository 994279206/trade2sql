package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Database 数据库接口
type Database struct {
	db     *sql.DB
	dbType string
}

// Connect 连接到数据库
func Connect(dbType, connStr string) (*Database, error) {
	db, err := sql.Open(dbType, connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &Database{
		db:     db,
		dbType: dbType,
	}, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	return d.db.Close()
}

// GetTableInfo 获取表结构信息
func (d *Database) GetTableInfo(tableName string) ([]ColumnInfo, error) {
	switch d.dbType {
	case "mysql":
		return d.getMySQLTableInfo(tableName)
	case "postgres":
		return d.getPostgresTableInfo(tableName)
	case "sqlite3":
		return d.getSQLiteTableInfo(tableName)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", d.dbType)
	}
}

// ColumnInfo 列信息
type ColumnInfo struct {
	Name       string
	Type       string
	IsNullable bool
	IsPrimary  bool
	Comment    string
}

// getMySQLTableInfo 获取MySQL表结构
func (d *Database) getMySQLTableInfo(tableName string) ([]ColumnInfo, error) {
	query := `
		SELECT 
			COLUMN_NAME, 
			DATA_TYPE, 
			IS_NULLABLE, 
			COLUMN_KEY, 
			COLUMN_COMMENT
		FROM 
			INFORMATION_SCHEMA.COLUMNS 
		WHERE 
			TABLE_NAME = ? 
		ORDER BY 
			ORDINAL_POSITION
	`

	rows, err := d.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var isNullable, columnKey string
		err := rows.Scan(&col.Name, &col.Type, &isNullable, &columnKey, &col.Comment)
		if err != nil {
			return nil, err
		}

		col.IsNullable = isNullable == "YES"
		col.IsPrimary = columnKey == "PRI"
		columns = append(columns, col)
	}

	return columns, nil
}

// getPostgresTableInfo 获取PostgreSQL表结构
func (d *Database) getPostgresTableInfo(tableName string) ([]ColumnInfo, error) {
	query := `
		SELECT 
			a.attname AS column_name,
			format_type(a.atttypid, a.atttypmod) AS data_type,
			CASE WHEN a.attnotnull THEN 'NO' ELSE 'YES' END AS is_nullable,
			CASE WHEN p.contype = 'p' THEN 'PRI' ELSE '' END AS column_key,
			(SELECT pg_catalog.col_description(a.attrelid, a.attnum)) AS column_comment
		FROM 
			pg_catalog.pg_attribute a
		LEFT JOIN 
			pg_catalog.pg_constraint p ON p.conrelid = a.attrelid AND a.attnum = ANY(p.conkey)
		WHERE 
			a.attrelid = $1::regclass
			AND a.attnum > 0
			AND NOT a.attisdropped
		ORDER BY 
			a.attnum
	`

	rows, err := d.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var isNullable, columnKey string
		err := rows.Scan(&col.Name, &col.Type, &isNullable, &columnKey, &col.Comment)
		if err != nil {
			return nil, err
		}

		col.IsNullable = isNullable == "YES"
		col.IsPrimary = columnKey == "PRI"
		columns = append(columns, col)
	}

	return columns, nil
}

// getSQLiteTableInfo 获取SQLite表结构
func (d *Database) getSQLiteTableInfo(tableName string) ([]ColumnInfo, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var dfltValue interface{}

		err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk)
		if err != nil {
			return nil, err
		}

		col := ColumnInfo{
			Name:       name,
			Type:       dataType,
			IsNullable: notNull == 0,
			IsPrimary:  pk > 0,
			Comment:    "",
		}
		columns = append(columns, col)
	}

	return columns, nil
}

// GetTableList 获取数据库中的所有表
func (d *Database) GetTableList() ([]string, error) {
	switch d.dbType {
	case "mysql":
		return d.getMySQLTableList()
	case "postgres":
		return d.getPostgresTableList()
	case "sqlite3":
		return d.getSQLiteTableList()
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", d.dbType)
	}
}

// getMySQLTableList 获取MySQL数据库中的所有表
func (d *Database) getMySQLTableList() ([]string, error) {
	// 获取当前数据库名
	var dbName string
	err := d.db.QueryRow("SELECT DATABASE()").Scan(&dbName)
	if err != nil {
		return nil, err
	}

	// 查询所有表
	query := `
		SELECT 
			TABLE_NAME 
		FROM 
			INFORMATION_SCHEMA.TABLES 
		WHERE 
			TABLE_SCHEMA = ? 
		ORDER BY 
			TABLE_NAME
	`

	rows, err := d.db.Query(query, dbName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

// getPostgresTableList 获取PostgreSQL数据库中的所有表
func (d *Database) getPostgresTableList() ([]string, error) {
	query := `
		SELECT 
			tablename 
		FROM 
			pg_catalog.pg_tables 
		WHERE 
			schemaname != 'pg_catalog' 
			AND schemaname != 'information_schema' 
		ORDER BY 
			tablename
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

// getSQLiteTableList 获取SQLite数据库中的所有表
func (d *Database) getSQLiteTableList() ([]string, error) {
	query := `
		SELECT 
			name 
		FROM 
			sqlite_master 
		WHERE 
			type='table' 
			AND name NOT LIKE 'sqlite_%' 
		ORDER BY 
			name
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}