package subdir

import "errors"

func foo() error {
	return errors.New("bar")
}
