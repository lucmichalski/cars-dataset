package admin

import (
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-contrib/motorcycles.autotrader.com/models"
)

const menuName = "motorcycles.autotrader.com"

// ConfigureAdmin configure admin interface
func ConfigureAdmin(Admin *admin.Admin) {

	Admin.AddMenu(&admin.Menu{Name: menuName, Priority: 1})

	// Add Setting page
	Admin.AddResource(&models.SettingAutoTraderMotorcycles{}, &admin.Config{
		Name:      menuName + " Settings",
		Menu:      []string{menuName},
		Singleton: true,
		Priority:  1,
	})

}