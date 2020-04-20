module github.com/lucmichalski/cars-contrib/carsdirect.com

replace github.com/lucmichalski/cars-dataset => ../..

go 1.14

require (
	github.com/PuerkitoBio/goquery v1.5.0
	github.com/cavaliercoder/grab v2.0.0+incompatible // indirect
	github.com/corpix/uarand v0.1.1
	github.com/gocolly/colly/v2 v2.0.1
	github.com/h2non/filetype v1.0.12 // indirect
	github.com/jinzhu/gorm v1.9.12
	github.com/k0kubun/pp v3.0.1+incompatible
	github.com/lucmichalski/cars-dataset v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	github.com/qor/admin v0.0.0-20200315024928-877b98a68a6f
	github.com/qor/media v0.0.0-20191022071353-19cf289e17d4
	github.com/qor/qor v0.0.0-20200224122013-457d2e3f50e1
	github.com/sirupsen/logrus v1.5.0
	github.com/tsak/concurrent-csv-writer v0.0.0-20200206204244-84054e222625 // indirect
)
