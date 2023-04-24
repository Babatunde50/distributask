package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type OperationType string

const (
	Resize           OperationType = "resize"
	Crop             OperationType = "crop"
	Rotate           OperationType = "rotate"
	Flip             OperationType = "flip"
	BrightnessAdjust OperationType = "brightness_adjust"
)

// ResizeParams holds the parameters for the Resize operation
type ResizeParams struct {
	Width        int  `json:"width"`
	Height       int  `json:"height"`
	Proportional bool `json:"proportional,omitempty"`
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
	Angle float64 `json:"angle"`
}

// BrightnessAdjustParams holds the parameters for the BrightnessAdjust operation
type BrightnessAdjustParams struct {
	Adjustment float64 `json:"adjustment"`
}

type Payload struct {
	URL       string        `json:"url"`
	Operation OperationType `json:"operation"`
	Params    interface{}   `json:"params,omitempty"`
}

func (p Payload) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Payload) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &p)
}
