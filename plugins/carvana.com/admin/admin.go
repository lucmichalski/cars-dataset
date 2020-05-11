package admin

import (
	"github.com/lucmichalski/cars-contrib/autosphere.fr/models"
	"github.com/qor/admin"
)

const menuName = "autosphere.fr"

// ConfigureAdmin configure admin interface
func ConfigureAdmin(Admin *admin.Admin) {

	Admin.AddMenu(&admin.Menu{Name: menuName, Priority: 1})

	// Add Setting page
	Admin.AddResource(&models.SettingAutosphere{}, &admin.Config{
		Name:      menuName + " Settings",
		Menu:      []string{menuName},
		Singleton: true,
		Priority:  1,
	})

}
