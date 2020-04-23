#---* Makefile *---#
.SILENT :

export GO111MODULE=on

# Base package
BASE_PACKAGE=github.com/lucmichalski

# App name
APPNAME=cars-dataset

# Go configuration
GOOS?=$(shell go env GOHOSTOS)
GOARCH?=$(shell go env GOHOSTARCH)

# Add exe extension if windows target
is_windows:=$(filter windows,$(GOOS))
EXT:=$(if $(is_windows),".exe","")
LDLAGS_LAUNCHER:=$(if $(is_windows),-ldflags "-H=windowsgui",)

# Archive name
ARCHIVE=$(APPNAME)-$(GOOS)-$(GOARCH).tgz

# Plugin name
PLUGIN?=cars-dataset

# Plugin filename
PLUGIN_SO=$(APPNAME)-$(PLUGIN).so

# Extract version infos
VERSION:=`git describe --tags --always`
GIT_COMMIT:=`git rev-list -1 HEAD --abbrev-commit`
BUILT:=`date`

## plugin				:	Build plugin (defined by PLUGIN variable).
plugin:
	mkdir -p release
	echo ">>> Building: $(PLUGIN_SO) $(VERSION) for $(GOOS)-$(GOARCH) ..."
	cd plugins/$(PLUGIN) && GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o ../../release/$(PLUGIN_SO)
.PHONY: plugin

## plugins			:	Build all qorpress plugins
plugins:
	GOARCH=amd64 PLUGIN=autoscout24.be make plugin
	GOARCH=amd64 PLUGIN=autoscout24.fr make plugin
	# GOARCH=amd64 PLUGIN=classics.autotrader.com.v2 make plugin
	# GOARCH=amd64 PLUGIN=motorcycles.autotrader.com.v2 make plugin
# 	GOARCH=amd64 PLUGIN=buyacar.co.uk make plugin
# 	GOARCH=amd64 PLUGIN=yamaha-occasion.com make plugin
# 	GOARCH=amd64 PLUGIN=motorcycles.autotrader.com make plugin
# 	GOARCH=amd64 PLUGIN=classiccars.com make plugin
# 	GOARCH=amd64 PLUGIN=classics.autotrader.com make plugin
# 	GOARCH=amd64 PLUGIN=carsdirect.com make plugin
# 	GOARCH=amd64 PLUGIN=cars.com make plugin
# 	GOARCH=amd64 PLUGIN=stanford-cars make plugin
# 	GOARCH=amd64 PLUGIN=carvana-kaggle make plugin
# 	GOARCH=amd64 PLUGIN=autosphere.fr make plugin
	# GOARCH=amd64 PLUGIN=autotrader.co.uk make plugin
	# GOARCH=amd64 PLUGIN=carvana.com make plugin
.PHONY: plugins  

## help				:	Print commands help.
help : Makefile
	@sed -n 's/^##//p' $<
.PHONY: help

# https://stackoverflow.com/a/6273809/1826109
%:
	@:
