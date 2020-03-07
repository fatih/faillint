//lint:file-ignore faillint ignore faillint in this file
package j

import (
	"errors"
)

func foo() error {
	return errors.New("bar!")
}
