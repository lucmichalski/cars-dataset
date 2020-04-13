module github.com/lucmichalski/cars-contrib/thecarconnection.com

replace github.com/lucmichalski/cars-dataset => ../..

go 1.14

require (
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/cavaliercoder/grab v2.0.0+incompatible // indirect
	github.com/corpix/uarand v0.1.1
	github.com/daetal-us/getld v1.0.1 // indirect
	github.com/gocolly/colly/v2 v2.0.1
	github.com/jinzhu/gorm v1.9.12
	github.com/k0kubun/pp v3.0.1+incompatible
	github.com/lucmichalski/cars-dataset v0.0.0-00010101000000-000000000000
	github.com/qor/media v0.0.0-20191022071353-19cf289e17d4
	github.com/sirupsen/logrus v1.5.0
	github.com/tsak/concurrent-csv-writer v0.0.0-20200206204244-84054e222625 // indirect
)
