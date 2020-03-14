// n is a package.
//
//lint:file-ignore faillint which is on the package comment but not the only comment.
package n

import (
	"errors"
)

func foo() error {
	return errors.New("bar!")
}
