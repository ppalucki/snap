/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package serror

import (
	"errors"
	"runtime"

	pkgerrors "github.com/pkg/errors"
)

type SnapError interface {
	error
	Fields() map[string]interface{}
	SetFields(map[string]interface{})
}

type Fields map[string]interface{}

type snapError struct {
	err    error
	fields Fields
}

// New returns an initialized SnapError.
// The variadic signature allows fields to optionally
// be added at construction.
// Can accept string or wrap and remembers location for better stacktraces.
func New(e interface{}, fields ...map[string]interface{}) *snapError {

	switch v := e.(type) {
	case error:
		// wrap functionallity
		pc, _, _, _ := runtime.Caller(1)
		e = struct {
			cause
			location
		}{
			cause{
				cause: v,
				// message: fmt.Sprintf("%T", e),
				message: "cause",
			},
			location(pc),
		}

	case string:
		// new
		// mimics errors.New API (accepts string)
		pc, _, _, _ := runtime.Caller(1)
		e = struct {
			error
			location
		}{
			errors.New(v),
			location(pc),
		}
	case SnapError:
		// Catch someone trying to wrap a serror around a serror.
		// We throw a panic to make them fix this.
		panic("You are trying to wrap a snapError around a snapError. Don't do this.")
	default:
		panic("can only accept bare 'error' and string nd SnapError types")
	}

	p := &snapError{
		err:    e.(error),
		fields: make(map[string]interface{}),
	}

	// insert fields into new snapError
	for _, f := range fields {
		for k, v := range f {
			p.fields[k] = v
		}
	}

	return p
}

func (p *snapError) SetFields(f map[string]interface{}) {
	p.fields = f
}

func (p *snapError) Fields() map[string]interface{} {
	return p.fields
}

func (p *snapError) Error() string {
	return p.err.Error()
}

func (p *snapError) String() string {
	return p.Error()
}

/*
Adding context to an error based on github.com/pkg/errors
*/

var (
	Wrap  = pkgerrors.Wrap
	Wrapf = pkgerrors.Wrapf
	Cause = pkgerrors.Cause
	Print = pkgerrors.Print
)

// shamefull copy-paste to have real location (stack unwinding)
type location uintptr

type cause struct {
	cause   error
	message string
}

func (c cause) Error() string   { return c.Message() + ": " + c.Cause().Error() }
func (c cause) Cause() error    { return c.cause }
func (c cause) Message() string { return c.message }
