package g

import (
	"errors"
	"fmt"
)

func foo() error {
	_ = fmt.Errorf("err") // want `function "Errorf" from package "fmt" shouldn't be used, suggested: "github.com/pkg/errors.{Errorf}"`
	fmt.Println("err")    // want `function "Println" from package "fmt" shouldn't be used`
	fmt.Print("err")      // want `function "Print" from package "fmt" shouldn't be used`
	_ = fmt.Sprintf("ok")

	// Second usage.
	_ = fmt.Errorf("err") // want `function "Errorf" from package "fmt" shouldn't be used, suggested: "github.com/pkg/errors.{Errorf}"`
	_ = fmt.Sprintf("ok")

	// More complex.
	func() {
		fmt.Print("err") // want `function "Print" from package "fmt" shouldn't be used`
	}()
	// Even more.
	defer func() {
		_ = fmt.Sprintf("%v", fmt.Errorf("err")) // want `function "Errorf" from package "fmt" shouldn't be used, suggested: "github.com/pkg/errors.{Errorf}"`
	}()
	return errors.New("bar!")
}
