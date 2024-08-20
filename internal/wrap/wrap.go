package wrap

import "errors"

func E(msg string, err error) error {
	return errors.New(msg + ": " + err.Error())
}
