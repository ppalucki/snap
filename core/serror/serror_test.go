package serror

import (
	"fmt"
	"log"
	"testing"
)

type SpecialError struct{}

func (e SpecialError) Error() string { return "specialdata" }

func fail() error {
	return SpecialError{}
}

func TestJustWrapping(t *testing.T) {
	a := func() error {
		err := fail()
		// return New(err, map[string]interface{}{"saya": "a-said"})
		return Wrap(err, "a-ctx")
		// return err
	}

	b := func() error {
		err := a()
		// return errors.New(err, map[string]interface{}{"sayb": "b-said"})
		return Wrap(err, "b-ctx")
	}

	c := func() error {
		err := b()
		// return errors.New(err, map[string]interface{}{"sayc": "c-said"})
		return Wrap(err, "c-ctx")
	}

	err := c()
	println("---- debug version - uses Print")
	Print(err)

	println()
	println("---- normal log version - handle by printing")
	log.Println(err)

	println()
	println("---- error unwrapping")
	switch v := Cause(err).(type) {
	case SnapError:
		fmt.Printf("snapError = %s fields=%v\n", v, v.Fields())
	case SpecialError:
		fmt.Printf("real cause: %T (%s)\n", v, v)
	default:
		fmt.Printf("other error: %T (%s)\n", v, v)
	}
}

func TestNewWrapping(t *testing.T) {
	a := func() error {
		err := fail()
		return New(err, map[string]interface{}{"saya": "a-said"})
	}

	b := func() error {
		err := a()
		return New(err, map[string]interface{}{"sayb": "b-said"})
	}

	c := func() error {
		err := b()
		return New(err, map[string]interface{}{"sayc": "c-said"})
	}

	err := c()
	println("---- debug version - uses Print")
	Print(err)

	println()
	println("---- normal log version - handle by printing")
	log.Println(err)

	println()
	println("---- error unwrapping")
	switch v := Cause(err).(type) {
	case SnapError:
		fmt.Printf("snapError = %s fields=%v\n", v, v.Fields())
	case SpecialError:
		fmt.Printf("real cause: %T (%s)\n", v, v)
	default:
		fmt.Printf("other error: %T (%s)\n", v, v)
	}
}
