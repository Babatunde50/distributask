package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type OperationType string

const (
	Resize OperationType = "resize"
	Crop   OperationType = "crop"
	Rotate OperationType = "rotate"
	Flip   OperationType = "flip"
)

// ResizeParams holds the parameters for the Resize operation
type ResizeParams struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// CropParams holds the parameters for the Crop operation
type CropParams struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// RotateParams holds the parameters for the Rotate operation
type RotateParams struct {
	Angle int `json:"angle"`
}

// BrightnessAdjustParams holds the parameters for the BrightnessAdjust operation

type FlipParams struct {
	Axis string `json:"axis"`
}

type TransformParams struct {
	ResizeParams ResizeParams
	CropParams   CropParams
	RotateParams RotateParams
	FlipParams   FlipParams
}

type Payload struct {
	URL       string          `json:"url"`
	Operation OperationType   `json:"operation"`
	Params    TransformParams `json:"params,omitempty"`
}

func (p Payload) Value() (driver.Value, error) {
	return json.Marshal(p)
}

type AllPossibleParams struct {
	Width  int    `json:"width,omitempty"`
	Axis   string `json:"axis,omitempty"`
	Angle  int    `json:"angle,omitempty"`
	X      int    `json:"x,omitempty"`
	Y      int    `json:"y,omitempty"`
	Height int    `json:"height,omitempty"`
}

func (p *Payload) UpdateParams(op OperationType, params AllPossibleParams) error {
	switch op {
	case Resize:
		if params.Width == 0 || params.Height == 0 {
			return errors.New("provide a width and height to resize image to")
		}
		p.Params.ResizeParams = ResizeParams{
			Width:  params.Width,
			Height: params.Height,
		}
	case Crop:
		if params.X == 0 || params.Y == 0 || params.Width == 0 || params.Height == 0 {
			return errors.New("provide a width,height,x, and y axis to crop image into")
		}
		p.Params.CropParams = CropParams{
			X:      params.X,
			Y:      params.Y,
			Width:  params.Width,
			Height: params.Height,
		}
	case Rotate:
		if params.Angle == 0 {
			return errors.New("provide an angle to rotate from")
		}
		p.Params.RotateParams = RotateParams{
			Angle: params.Angle,
		}
	case Flip:
		p.Params.FlipParams = FlipParams{
			Axis: params.Axis,
		}
	}

	return nil

}

func (p *Payload) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &p)
}
