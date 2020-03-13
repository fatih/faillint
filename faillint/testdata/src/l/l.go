package l

import (
	"errors"
)

func foo() error {
	//lint:ignore faillint ignore this errors.New usage
	return errors.New("bar!")
}
