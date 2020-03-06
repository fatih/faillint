package h

import (
	"errors" // want `package "errors" shouldn't be imported`
)

func foo() error {
	return errors.New("bar!")
}
