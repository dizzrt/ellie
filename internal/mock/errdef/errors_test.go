package errdef

import (
	"errors"
	"fmt"
	"testing"
)

func TestStandardError_Error(t *testing.T) {
	ee := errors.New("test error")

	err := InvalidParams().WithMessage("invalid params message").WithCause(ee).WithMetadata(map[string]string{
		"key": "value",
	})
	fmt.Println(err.Error())

	err2 := InvalidParams().WithCause(ee)
	fmt.Println(err2.Error())

	err3 := InvalidParams().WithCause(err2)
	fmt.Println(err3.Error())
}
