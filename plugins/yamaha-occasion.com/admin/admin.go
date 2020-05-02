package admin

import (
	"github.com/lucmichalski/cars-contrib/yamaha-occasion.com/models"
	"github.com/qor/admin"
)

const menuName = "yamaha-occasion.com"

// ConfigureAdmin configure admin interface
func ConfigureAdmin(Admin *admin.Admin) {

	Admin.AddMenu(&admin.Menu{Name: menuName, Priority: 1})

	// Add Setting page
	Admin.AddResource(&models.SettingYamahaOccasion{}, &admin.Config{
		Name:      menuName + " Settings",
		Menu:      []string{menuName},
		Singleton: true,
		Priority:  1,
	})

}
