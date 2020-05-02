package admin

import (
	"github.com/lucmichalski/cars-contrib/autoscout24.be/models"
	"github.com/qor/admin"
)

const menuName = "autoscout24.be"

// ConfigureAdmin configure admin interface
func ConfigureAdmin(Admin *admin.Admin) {

	Admin.AddMenu(&admin.Menu{Name: menuName, Priority: 1})

	// Add Setting page
	Admin.AddResource(&models.SettingAutoScout24Be{}, &admin.Config{
		Name:      menuName + " Settings",
		Menu:      []string{menuName},
		Singleton: true,
		Priority:  1,
	})

}
