package worker

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Babatunde50/distributask/internal/database"
	"github.com/h2non/bimg"
)

func doTask(dbTask *database.Task, db *database.DB, fn func() ([]byte, error)) error {

	updatedImage, err := fn()

	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(updatedImage)

	dbTask.Result = encoded

	dbTask.Status = "completed"

	err = db.UpdateTask(dbTask)

	if err != nil {
		return fmt.Errorf("error updating task: %v", err)
	}

	return nil
}

func imageHandler(db *database.DB, dbTask *database.Task) error {
	imageURL := dbTask.Payload.URL

	res, err := http.Get(imageURL)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return fmt.Errorf("failed to read image: %v", err)
	}

	switch dbTask.Payload.Operation {
	case database.Resize:
		return doTask(dbTask, db, func() ([]byte, error) {
			resizedImage, err := bimg.NewImage(data).Resize(dbTask.Payload.Params.ResizeParams.Width, dbTask.Payload.Params.ResizeParams.Height)

			if err != nil {
				return nil, fmt.Errorf("error resizing image: %v", err)
			}

			return resizedImage, nil
		})

	case database.Crop:
		return doTask(dbTask, db, func() ([]byte, error) {
			croppedImage, err := bimg.NewImage(data).Crop(dbTask.Payload.Params.CropParams.X, dbTask.Payload.Params.CropParams.Y, bimg.GravityCentre)

			if err != nil {
				return nil, fmt.Errorf("error cropping image: %v", err)
			}

			return croppedImage, nil
		})

	case database.Flip:
		return doTask(dbTask, db, func() ([]byte, error) {
			var flippedImage []byte
			var err error
			if dbTask.Payload.Params.FlipParams.Axis == "X" {
				flippedImage, err = bimg.NewImage(data).Flop()
			} else {
				flippedImage, err = bimg.NewImage(data).Flip()
			}

			if err != nil {
				return nil, fmt.Errorf("error flipping image: %v", err)
			}

			return flippedImage, nil
		})
	case database.Rotate:
		return doTask(dbTask, db, func() ([]byte, error) {
			rotatedImage, err := bimg.NewImage(data).Rotate(bimg.Angle(dbTask.Payload.Params.RotateParams.Angle))

			if err != nil {
				return nil, fmt.Errorf("error cropping image: %v", err)
			}

			return rotatedImage, nil
		})
	default:
		return fmt.Errorf("unimplemented operation: %v", dbTask.Payload.Operation)
	}

}
