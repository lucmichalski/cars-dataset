package models

type Labelme struct {
	FillColor   []int       `json:"fillColor"`
	Flags       Flags       `json:"flags"`
	ImageData   string 		`json:"imageData"`
	ImageHeight int         `json:"imageHeight"`
	ImagePath   string      `json:"imagePath"`
	ImageWidth  int         `json:"imageWidth"`
	LineColor   []int       `json:"lineColor"`
	Shapes      []Shape     `json:"shapes"`
	Version     string      `json:"version"`
}

type Flags struct {
}

type Shape struct {
	FillColor []int 	  `json:"fill_color"`
	Label     string      `json:"label"`
	LineColor []int 	  `json:"line_color"`
	Points    [][]int     `json:"points"`
	ShapeType string      `json:"shape_type"`
}