module github.com/lucmichalski/cars-contrib/leboncoin.fr

replace github.com/lucmichalski/cars-dataset => ../..

go 1.14

require (
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/k0kubun/pp v3.0.1+incompatible
	github.com/lucmichalski/cars-dataset v0.0.0-00010101000000-000000000000
	github.com/nozzle/throttler v0.0.0-20180817012639-2ea982251481
	github.com/sirupsen/logrus v1.5.0
	github.com/tsak/concurrent-csv-writer v0.0.0-20200206204244-84054e222625 // indirect
)
