package e

import (
	"errors"                   // want `package "errors" shouldn't be imported, suggested: "github.com/pkg/errors"`
	"golang.org/x/net/context" // want `package "golang.org/x/net/context" shouldn't be imported, suggested: "context"`
)

func foo(ctx context.Context) error {
	return errors.New("bar!")
}
