package admin

import (
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-contrib/autosphere.fr/models"
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
