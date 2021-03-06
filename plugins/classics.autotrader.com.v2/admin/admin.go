package admin

import (
	"github.com/lucmichalski/cars-contrib/classics.autotrader.com.v2/models"
	"github.com/qor/admin"
)

const menuName = "classics.autotrader.com.v2"

// ConfigureAdmin configure admin interface
func ConfigureAdmin(Admin *admin.Admin) {

	Admin.AddMenu(&admin.Menu{Name: menuName, Priority: 1})

	// Add Setting page
	Admin.AddResource(&models.SettingClassicAutoTraderV2{}, &admin.Config{
		Name:      menuName + " Settings",
		Menu:      []string{menuName},
		Singleton: true,
		Priority:  1,
	})

}
