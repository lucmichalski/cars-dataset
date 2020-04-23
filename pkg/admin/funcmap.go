package admin

import (
	"html/template"

	"github.com/qor/admin"
)

func initFuncMap(Admin *admin.Admin) {
	Admin.RegisterFuncMap("render_latest_vehicles", renderLatestVehicles)
}

func renderLatestVehicles(context *admin.Context) template.HTML {
	var vehicleContext = context.NewResourceContext("Vehicle")
	vehicleContext.Searcher.Pagination.PerPage = 5
	if products, err := vehicleContext.FindMany(); err == nil {
		return vehicleContext.Render("index/table", products)
	}
	return template.HTML("")
}
