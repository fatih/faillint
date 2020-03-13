//lint:file-ignore faillint which is on the package comment but not the only comment.

// o is a package.
package o

import (
	"errors"
)

func foo() error {
	return errors.New("bar!")
}
