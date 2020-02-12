package e

import (
	"golang.org/x/net/context" // want `package "golang.org/x/net/context" shouldn't be imported, suggested: "context"`
)

func foo(ctx context.Context) {
}
