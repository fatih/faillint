package l

import (
	"errors"
)

func foo() error {
	//faillint:ignore ignore this errors.New usage
	return errors.New("bar!")
}
