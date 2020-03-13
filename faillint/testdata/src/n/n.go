// n is a package.
//
//faillint:file-ignore which is on the package comment but not the only comment.
package n

import (
	"errors"
)

func foo() error {
	return errors.New("bar!")
}
