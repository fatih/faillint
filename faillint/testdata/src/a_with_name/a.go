package a

import (
	mislead "errors" // want `package "errors" shouldn't be imported`
)

func foo() error {
	return mislead.New("bar!")
}
