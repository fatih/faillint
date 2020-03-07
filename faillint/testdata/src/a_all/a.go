package a

import (
	. "errors" // want `package "errors" shouldn't be imported`
)

func foo() { New("bar!") }
