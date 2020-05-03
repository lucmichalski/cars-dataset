package admin

import (
	"html/template"

	"github.com/qor/admin"
	// "github.com/k0kubun/pp"
)

func initFuncMap(Admin *admin.Admin) {
	Admin.RegisterFuncMap("render_latest_vehicles", renderLatestVehicles)
	// Admin.RegisterFuncMap("render_latest_vehicle_images", renderLatestVehicleImages)
}

func renderLatestVehicles(context *admin.Context) template.HTML {
	var vehicleContext = context.NewResourceContext("Vehicle")
	vehicleContext.Searcher.Pagination.PerPage = 25
	if vehicles, err := vehicleContext.FindMany(); err == nil {
		return vehicleContext.Render("index/table", vehicles)
	}
	return template.HTML("")
}

/*
func renderLatestVehicleImages(context *admin.Context) template.HTML {
	var vehicleImagesContext = context.NewResourceContext("Vehicle Images")
	pp.Println(vehicleImagesContext)
	vehicleImagesContext.Searcher.Pagination.PerPage = 5
	if vehicleImages, err := vehicleImagesContext.FindMany(); err == nil {
		return vehicleImagesContext.Render("index/table", vehicleImages)
	}
	return template.HTML("")
}
*/
