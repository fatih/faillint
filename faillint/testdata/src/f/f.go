package e

import (
	"golang.org/x/net/context" // want `sub-package of "golang.org/x/net" shouldn't be imported`
)

func foo(ctx context.Context) {
}
