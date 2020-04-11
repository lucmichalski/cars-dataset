package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/media/media_library"
	"github.com/qor/validations"
)

type Vehicle struct {
	gorm.Model
	URL               string `gorm:"index:url"`
	Name              string `gorm:"index:name"`
	Modl              string `gorm:"index:modl"`
	Engine            string `gorm:"index:engine"`
	Year              string `gorm:"index:year"`
	Source            string `gorm:"index:source"`
	Gid               string `gorm:"index:gid"`
	Manufacturer      string `gorm:"index:manufacturer"`
	MainImage         media_library.MediaBox
	Images            media_library.MediaBox
	VehicleProperties VehicleProperties `sql:"type:text"`
}

func (v Vehicle) MainImageURL(styles ...string) string {
	style := "original"
	if len(styles) > 0 {
		style = styles[0]
	}

	if len(v.MainImage.Files) > 0 {
		return v.MainImage.URL(style)
	}
	return "/images/no_image.png"
}

func (v Vehicle) Validate(db *gorm.DB) {
	if strings.TrimSpace(v.Name) == "" {
		db.AddError(validations.NewError(v, "Name", "Name can not be empty"))
	}
}

func (v *Vehicle) AfterCreate() (err error) {
	// add to manticore
	// add to bleve
	return
}

type VehicleImage struct {
	gorm.Model
	Title        string
	Checksum     string
	SelectedType string
	File         media_library.MediaLibraryStorage `sql:"size:4294967295;" media_library:"url:/system/{{class}}/{{primary_key}}/{{column}}.{{extension}}"`
}

func (vehicleImage VehicleImage) Validate(db *gorm.DB) {
	if strings.TrimSpace(vehicleImage.Title) == "" {
		db.AddError(validations.NewError(vehicleImage, "Title", "Title can not be empty"))
	}
}

func (vehicleImage *VehicleImage) SetSelectedType(typ string) {
	vehicleImage.SelectedType = typ
}

func (vehicleImage *VehicleImage) GetSelectedType() string {
	return vehicleImage.SelectedType
}

func (vehicleImage *VehicleImage) ScanMediaOptions(mediaOption media_library.MediaOption) error {
	if bytes, err := json.Marshal(mediaOption); err == nil {
		return vehicleImage.File.Scan(bytes)
	} else {
		return err
	}
}

func (vehicleImage *VehicleImage) GetMediaOption() (mediaOption media_library.MediaOption) {
	mediaOption.Video = vehicleImage.File.Video
	mediaOption.FileName = vehicleImage.File.FileName
	mediaOption.URL = vehicleImage.File.URL()
	mediaOption.OriginalURL = vehicleImage.File.URL("original")
	mediaOption.CropOptions = vehicleImage.File.CropOptions
	mediaOption.Sizes = vehicleImage.File.GetSizes()
	mediaOption.Description = vehicleImage.File.Description
	return
}

type VehicleProperties []VehicleProperty

type VehicleProperty struct {
	Name  string
	Value string
}

func (vehicleProperties *VehicleProperties) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, vehicleProperties)
	case string:
		if v != "" {
			return vehicleProperties.Scan([]byte(v))
		}
	default:
		return errors.New("not supported")
	}
	return nil
}

func (vehicleProperties VehicleProperties) Value() (driver.Value, error) {
	if len(vehicleProperties) == 0 {
		return nil, nil
	}
	return json.Marshal(vehicleProperties)
}
