package i

import (
	"os"
)

func foo() {
	// Constants.
	_ = os.O_RDONLY // want `declaration "O_RDONLY" from package "os" shouldn't be used`
	_ = os.O_CREATE

	// Vars.
	_ = os.ErrNotExist // want `declaration "ErrNotExist" from package "os" shouldn't be used`
	_ = os.ErrClosed

	// Types.
	type _ os.File // want `declaration "File" from package "os" shouldn't be used`
	type _ os.FileInfo
}
