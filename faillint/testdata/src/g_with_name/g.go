package g

import (
	"errors"
	"fmt"
	mislead "fmt"
)

func foo() error {
	_ = mislead.Errorf("err") // want `function "Errorf" from package "fmt" shouldn't be used, suggested: "github.com/pkg/errors.{Errorf}"`
	mislead.Println("err")    // want `function "Println" from package "fmt" shouldn't be used`
	fmt.Print("err")          // want `function "Print" from package "fmt" shouldn't be used`
	_ = fmt.Sprintf("ok")
	return errors.New("bar!")
}
