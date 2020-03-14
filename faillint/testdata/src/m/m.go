package m

import (
	"errors"
)

func foo() error {
	_ = errors.New("foo!") // want `declaration "New" from package "errors" shouldn't be used`

	_ = errors.New("baz!") //lint:ignore faillint ignore this errors.New usage

	//lint:ignore faillint ignore this errors.New usage
	return errors.New("bar!")
}
