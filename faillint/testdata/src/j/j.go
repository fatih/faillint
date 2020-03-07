package j

import (
	//faillint:ignore tolerate this errors import
	"errors"
	"fmt" //faillint:ignore tolerate this fmt import
)

func foo() error {
	return errors.New(fmt.Sprintf("%s!", "bar"))
}
