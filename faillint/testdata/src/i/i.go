package i

import (
	//lint:ignore faillint tolerate this errors
	"errors"
)

func foo() error {
	return errors.New("bar!")
}
