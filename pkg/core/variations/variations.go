package variations

import (
	"errors"

	"github.com/lucmichalski/cars-dataset/pkg/core/admin"
	"github.com/lucmichalski/cars-dataset/pkg/core/qor/resource"
	"github.com/lucmichalski/cars-dataset/pkg/core/qor/utils"
)

func init() {
	admin.RegisterViewPath("github.com/lucmichalski/cars-dataset/pkg/core/variations/views")
}

type VariationsConfig struct {
	resource     *admin.Resource
	PrimaryAttrs []string // primary attrs should in EditAttrs of the variation resource
}

func (variationsConfig VariationsConfig) PrimaryAttributeMetas() (metas []*admin.Meta) {
	for _, name := range variationsConfig.PrimaryAttrs {
		metas = append(metas, variationsConfig.resource.GetMeta(name))
	}
	return
}

func (variationsConfig *VariationsConfig) GetTemplate(context *admin.Context, metaType string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (variationsConfig *VariationsConfig) ConfigureQorMeta(metaor resource.Metaor) {
	if meta, ok := metaor.(*admin.Meta); ok {
		meta.Type = "variations"

		variationsRes := meta.Resource
		variationsConfig.resource = variationsRes

		if len(variationsConfig.PrimaryAttrs) == 0 {
			for _, meta := range variationsRes.ConvertSectionToMetas(meta.Resource.EditAttrs()) {
				if meta.FieldStruct != nil {
					tagSettings := utils.ParseTagOption(meta.FieldStruct.Tag.Get("variations"))
					if _, ok := tagSettings["PRIMARY"]; ok {
						variationsConfig.PrimaryAttrs = append(variationsConfig.PrimaryAttrs, meta.Name)
					}
				}
			}
		}
	}
}
