module github.com/lucmichalski/cars-contrib/autotrader.com

replace github.com/lucmichalski/cars-dataset => ../..

go 1.14

require (
	github.com/cavaliercoder/grab v2.0.0+incompatible // indirect
	github.com/jinzhu/gorm v1.9.12
	github.com/k0kubun/pp v3.0.1+incompatible
	github.com/lucmichalski/cars-dataset v0.0.0-00010101000000-000000000000
	github.com/qor/admin v0.0.0-20200315024928-877b98a68a6f
	github.com/sirupsen/logrus v1.5.0
	github.com/tsak/concurrent-csv-writer v0.0.0-20200206204244-84054e222625 // indirect
	github.com/x0rzkov/selenium v1.0.2
)
