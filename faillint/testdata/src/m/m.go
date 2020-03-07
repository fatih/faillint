package m

import (
	"errors"
)

func foo() error {
	_ = errors.New("foo!") // want `declaration "New" from package "errors" shouldn't be used`

	_ = errors.New("baz!") //faillint:ignore ignore this errors.New usage

	//faillint:ignore ignore this errors.New usage
	return errors.New("bar!")
}
