package h

import (
	"errors"
)

func fooTest() error {
	return errors.New("bar!")
}
