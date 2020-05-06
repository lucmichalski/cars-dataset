#!/bin/sh



nohup go run main.go --plugins autosphere.fr --extract --no-cache --clean --verbose > ./shared/nohup/autosphere.fr &
nohup go run main.go --plugins autoscout24.fr --extract --no-cache --clean --verbose > ./shared/nohup/autoscout24.fr &
nohup go run main.go --plugins autoscout24.be --extract --no-cache --clean --verbose > ./shared/nohup/autoscout24.be.out &
nohup go run main.go --plugins motorcycles.autotrader.com.v2 --extract --no-cache --clean --verbose > ./dhared/nohup/motorcycles.autotrader.com.v2.out &
cp shared/queue/cars.com_sitemap.txt.4 shared/queue/cars.com_sitemap.txt 
nohup go run main.go --plugins cars.com --extract --no-cache --clean --verbose > ./shared/nohup/cars.com.out &
nohup go run main.go --plugins thecarconnection.com --extract --no-cache --clean --verbose > ./shared/nohup/thecarconnection.com.out &
nohup go run main.go --plugins classics.autotrader.com.v2 --extract --no-cache --verbose --clean > ./shared/nohup/classics.autotrader.com.v2.out &
