package g

import (
	"errors"
	"fmt"
	mislead "fmt"
)

func foo() error {
	_ = mislead.Errorf("err") // want `declaration "Errorf" from package "fmt" shouldn't be used, suggested: "github.com/pkg/errors.{Errorf}"`
	mislead.Println("err")    // want `declaration "Println" from package "fmt" shouldn't be used`
	fmt.Print("err")          // want `declaration "Print" from package "fmt" shouldn't be used`
	_ = fmt.Sprintf("ok")
	return errors.New("bar!")
}
