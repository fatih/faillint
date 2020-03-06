package a

import (
	"errors"          // want `package "errors" shouldn't be imported`
	mislead "errors"  // want `package "errors" shouldn't be imported`
	mislead2 "errors" // want `package "errors" shouldn't be imported`
)

func foo() error {
	_ = errors.New("")
	_ = mislead2.New("")
	return mislead.New("bar!")
}
