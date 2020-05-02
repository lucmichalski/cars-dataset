package admin

import (
	"github.com/lucmichalski/cars-contrib/autoscout24.fr/models"
	"github.com/qor/admin"
)

const menuName = "autoscout24.fr"

// ConfigureAdmin configure admin interface
func ConfigureAdmin(Admin *admin.Admin) {

	Admin.AddMenu(&admin.Menu{Name: menuName, Priority: 1})

	// Add Setting page
	Admin.AddResource(&models.SettingAutoScout24Fr{}, &admin.Config{
		Name:      menuName + " Settings",
		Menu:      []string{menuName},
		Singleton: true,
		Priority:  1,
	})

}
