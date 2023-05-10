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

func (p *Payload) UpdateParams(op OperationType, params AllPossibleParams) {
	switch op {
	case Resize:
		// do something
		p.Params.ResizeParams = ResizeParams{
			Width:  params.Width,
			Height: params.Height,
		}
	case Crop:
		// do something
		p.Params.CropParams = CropParams{
			X:      params.X,
			Y:      params.Y,
			Width:  params.Width,
			Height: params.Height,
		}
	case Rotate:
		// do something
		p.Params.RotateParams = RotateParams{
			Angle: params.Angle,
		}
	case Flip:
		// do something
		p.Params.FlipParams = FlipParams{
			Axis: params.Axis,
		}
	}

}

func (p *Payload) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &p)
}
