package admin

import (
	"github.com/lucmichalski/cars-contrib/thecarconnection.com/models"
	"github.com/qor/admin"
)

const menuName = "thecarconnection.com"

// ConfigureAdmin configure admin interface
func ConfigureAdmin(Admin *admin.Admin) {

	Admin.AddMenu(&admin.Menu{Name: menuName, Priority: 1})

	// Add Setting page
	Admin.AddResource(&models.SettingTheCarConnection{}, &admin.Config{
		Name:      menuName + " Settings",
		Menu:      []string{menuName},
		Singleton: true,
		Priority:  1,
	})

}
