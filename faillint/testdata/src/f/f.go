package f

import (
	"errors"
	fake "github.com/fatih/faillint/faillint/testdata/src/fake.NotFunction" // want `package "github.com/fatih/faillint/faillint/testdata/src/fake.NotFunction" shouldn't be imported, suggested: "not_fake"`
)

func foo() error {
	fake.Fake()
	return errors.New("bar!")
}
