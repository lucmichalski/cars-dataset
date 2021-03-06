package admin

import (
	"github.com/lucmichalski/cars-contrib/buyacar.co.uk/models"
	"github.com/qor/admin"
)

const menuName = "buyacar.co.uk"

// ConfigureAdmin configure admin interface
func ConfigureAdmin(Admin *admin.Admin) {

	Admin.AddMenu(&admin.Menu{Name: menuName, Priority: 1})

	// Add Setting page
	Admin.AddResource(&models.SettingBuyACar{}, &admin.Config{
		Name:      menuName + " Settings",
		Menu:      []string{menuName},
		Singleton: true,
		Priority:  1,
	})

}
