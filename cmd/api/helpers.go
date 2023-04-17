package main

import (
	"fmt"
)

func (app *application) backgroundTask(fn func() error) {
	app.wg.Add(1)

	go func() {
		defer app.wg.Done()

		defer func() {
			err := recover()
			if err != nil {
				app.reportError(fmt.Errorf("%s", err))
			}
		}()

		err := fn()
		if err != nil {
			app.reportError(err)
		}
	}()
}
