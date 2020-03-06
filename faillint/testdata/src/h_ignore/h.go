package h

import (
	"errors" // want `package "errors" shouldn't be imported, suggested: "github.com/pkg/errors"`
)

func foo() error {
	return errors.New("bar!")
}
