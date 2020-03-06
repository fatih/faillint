package h_test

import (
	"errors" // want `package "errors" shouldn't be imported`
	"testing"
)

func TestFoo(t *testing.T) {
    t.Errorf("Got bar error: %g", errors.New("bar!"))
}
