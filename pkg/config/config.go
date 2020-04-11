package config

import (
	"database/sql"
	"os"

	"github.com/jinzhu/gorm"
)

type Config struct {
	IsDebug         bool
	IsSitemapIndex  bool
	AllowedDomains  []string
	CacheDir        string
	ConsumerThreads int
	QueueMaxSize    int
	URLs            []string
	DryMode         bool
	IsClean         bool
	AnalyzerURL     string
	CatalogURL      string
	ImageDir        string
	DumpDir         string   `default:"./shared/dump"`
	DB              *gorm.DB `gorm:"-"`
	IDX             *sql.DB  `gorm:"-"`
	Index           string
	Writer          *os.File
}
