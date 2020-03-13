package j

import (
	//lint:ignore faillint tolerate this errors import
	"errors"
	"fmt" //lint:ignore faillint tolerate this fmt import
)

func foo() error {
	return errors.New(fmt.Sprintf("%s!", "bar"))
}
