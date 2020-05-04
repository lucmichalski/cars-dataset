package daterange

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/lucmichalski/cars-dataset/pkg/core/admin"
	"github.com/lucmichalski/cars-dataset/pkg/core/qor/resource"
)

func init() {
	admin.RegisterViewPath("github.com/lucmichalski/cars-dataset/pkg/core/metas/daterange/views")
}

// DateRanger date ranger interface
type DateRanger interface {
	GetDateRange() *DateRange
}

// RangeType range type
type RangeType = string

var (
	// Today today range type
	Today RangeType = "today"
	// Yesterday yesterday
	Yesterday RangeType = "yesterday"
	// LastWeek last week
	LastWeek RangeType = "last_week"
	// LastMonth last month
	LastMonth RangeType = "last_month"
	// Last7Days last 7 days
	Last7Days RangeType = "last_7_days"
	// Last30Days last 30 days
	Last30Days RangeType = "last_30_days"
)

// DateRange date range type
type DateRange struct {
	Type  RangeType
	From  *time.Time
	Until *time.Time
}

// Scan scan date range value
func (dateRange *DateRange) Scan(value interface{}) error {
	switch data := value.(type) {
	case []byte:
		return json.Unmarshal(data, dateRange)
	case string:
		return dateRange.Scan([]byte(data))
	case []string:
		for _, str := range data {
			if err := dateRange.Scan([]byte(str)); err != nil {
				return err
			}
		}
	default:
		return errors.New("unsupported data")
	}
	return nil
}

// Value get value of dateRange
func (dateRange DateRange) Value() (driver.Value, error) {
	return json.Marshal(dateRange)
}

// ConfigureQorMeta configure qor meta
func (dateRange DateRange) ConfigureQorMeta(metaor resource.Metaor) {
	if meta, ok := metaor.(*admin.Meta); ok {
		meta.Type = "daterange"
	}
}

// GetDateRange get date range
func (dateRange DateRange) GetDateRange() *DateRange {
	return &dateRange
}
