package sleepintest

import (
	"testing"
	"time"
)

func TestFoo(t *testing.T) {
	time.Sleep(1 * time.Second) // want `declaration "Sleep" from package "time" shouldn't be used`
}
