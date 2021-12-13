package r

import "fmt" // want `package "fmt" shouldn't be imported`

func foo() error {
	return fmt.Errorf("")
}
