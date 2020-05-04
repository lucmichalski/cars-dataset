package publish2_test

import (
	"github.com/lucmichalski/cars-dataset/pkg/core/l10n"
	"github.com/lucmichalski/cars-dataset/pkg/core/publish2"
	"github.com/lucmichalski/cars-dataset/pkg/core/qor/test/utils"
)

var DB = utils.TestDB()

func init() {
	models := []interface{}{
		&Wiki{}, &Post{}, &Article{}, &Discount{}, &User{}, &Campaign{},
		&Product{}, &L10nProduct{}, &SharedVersionProduct{}, &SharedVersionColorVariation{}, &SharedVersionSizeVariation{},
	}

	DB.DropTableIfExists(models...)
	DB.AutoMigrate(models...)
	publish2.RegisterCallbacks(DB)
	l10n.RegisterCallbacks(DB)
}
