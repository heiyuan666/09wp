package database

import (
	"database/sql"
	"fmt"
	"log"

	"PanCheck/internal/model"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Config 数据库配置
type Config struct {
	Type     string `yaml:"type"` // mysql/postgres
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"` // mysql only
}

// Init 初始化数据库连接
func Init(config Config) error {
	// 先确保数据库存在
	if err := ensureDatabase(config); err != nil {
		return fmt.Errorf("failed to ensure database exists: %w", err)
	}

	var dialector gorm.Dialector
	var err error

	dsn := ""
	switch config.Type {
	case "mysql":
		if config.Charset == "" {
			config.Charset = "utf8mb4"
		}
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			config.User, config.Password, config.Host, config.Port, config.Database, config.Charset)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			config.Host, config.User, config.Password, config.Database, config.Port)
		dialector = postgres.Open(dsn)
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移
	err = AutoMigrate()
	if err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

// ensureDatabase 确保数据库存在，如果不存在则创建
func ensureDatabase(config Config) error {
	switch config.Type {
	case "mysql":
		return ensureMySQLDatabase(config)
	case "postgres":
		return ensurePostgresDatabase(config)
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// ensureMySQLDatabase 确保 MySQL 数据库存在
func ensureMySQLDatabase(config Config) error {
	// 连接到 MySQL 服务器（不指定数据库）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL server: %w", err)
	}
	defer db.Close()

	// 测试连接
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL server: %w", err)
	}

	// 检查数据库是否存在（使用参数化查询避免 SQL 注入）
	var exists int
	query := "SELECT 1 FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?"
	err = db.QueryRow(query, config.Database).Scan(&exists)
	if err == nil {
		// 数据库已存在
		log.Printf("Database '%s' already exists\n", config.Database)
		return nil
	}

	if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// 数据库不存在，创建它
	log.Printf("Database '%s' does not exist, creating...\n", config.Database)
	// MySQL 的 CREATE DATABASE 不支持参数化查询，所以需要对数据库名进行转义
	createSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", escapeMySQLIdentifier(config.Database))
	_, err = db.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	log.Printf("Database '%s' created successfully\n", config.Database)
	return nil
}

// escapeMySQLIdentifier 转义 MySQL 标识符
func escapeMySQLIdentifier(name string) string {
	// MySQL 使用反引号转义，需要将反引号替换为双反引号
	escaped := ""
	for _, r := range name {
		if r == '`' {
			escaped += "``"
		} else {
			escaped += string(r)
		}
	}
	return escaped
}

// ensurePostgresDatabase 确保 PostgreSQL 数据库存在
func ensurePostgresDatabase(config Config) error {
	// 连接到 PostgreSQL 服务器（连接到默认的 postgres 数据库）
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%d sslmode=disable TimeZone=Asia/Shanghai",
		config.Host, config.User, config.Password, config.Port)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL server: %w", err)
	}
	defer db.Close()

	// 测试连接
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL server: %w", err)
	}

	// 检查数据库是否存在
	var exists int
	query := "SELECT 1 FROM pg_database WHERE datname = $1"
	err = db.QueryRow(query, config.Database).Scan(&exists)
	if err == nil {
		// 数据库已存在
		log.Printf("Database '%s' already exists\n", config.Database)
		return nil
	}

	if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// 数据库不存在，创建它
	log.Printf("Database '%s' does not exist, creating...\n", config.Database)
	// PostgreSQL 不允许在事务中创建数据库，需要设置 autocommit
	// 使用参数化查询避免 SQL 注入，但 PostgreSQL 的 CREATE DATABASE 不支持参数化
	// 所以需要对数据库名进行转义
	createSQL := fmt.Sprintf("CREATE DATABASE %s", quotePostgresIdentifier(config.Database))
	_, err = db.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	log.Printf("Database '%s' created successfully\n", config.Database)
	return nil
}

// quotePostgresIdentifier 转义 PostgreSQL 标识符
func quotePostgresIdentifier(name string) string {
	// PostgreSQL 使用双引号转义，需要将双引号替换为双双引号
	escaped := ""
	for _, r := range name {
		if r == '"' {
			escaped += `""`
		} else {
			escaped += string(r)
		}
	}
	return fmt.Sprintf(`"%s"`, escaped)
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate() error {
	return DB.AutoMigrate(
		&model.SubmissionRecord{},
		&model.InvalidLink{},
		&model.Setting{},
		&model.ScheduledTask{},
		&model.TaskExecution{},
	)
}

// Close 关闭数据库连接
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
