package assert

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

type Assert struct {
	*testing.T
	name string
}

func NewAssert(t *testing.T) *Assert {
	return &Assert{T: t}
}

func NewAssertWithName(t *testing.T, name string) *Assert {
	return &Assert{t, name}
}

func (ast *Assert) Log(args ...interface{}) {
	if ast.name != "" {
		args = append(args, 0)
		copy(args[1:], args[0:])
		args[0] = ast.name + ":"
		ast.T.Log(args...)
	} else {
		ast.T.Log(args...)
	}
}

func (ast *Assert) Logf(format string, args ...interface{}) {
	if ast.name != "" {
		ast.T.Logf(ast.name+": "+format, args...)
	} else {
		ast.T.Logf(format, args...)
	}
}

func (ast *Assert) Nil(value interface{}, logs ...interface{}) {
	ast.nilAssert(true, true, value, logs...)
}

func (ast *Assert) NotNil(value interface{}, logs ...interface{}) {
	ast.nilAssert(true, false, value, logs...)
}

func isNil(value interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	k := v.Kind()

	switch k {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}
	return false
}

func (ast *Assert) nilAssert(fatal bool, expNil bool, value interface{}, logs ...interface{}) {
	if expNil != isNil(value) {
		ast.logCaller()
		if len(logs) > 0 {
			ast.Log(logs...)
		} else {
			if expNil {
				ast.Log("value is not nil:", value)
			} else {
				ast.Log("value is nil")
			}
		}
		ast.failIt(fatal)
	}
}

func (ast *Assert) True(boolValue bool, logs ...interface{}) {
	ast.trueAssert(true, boolValue, logs...)
}

func (ast *Assert) trueAssert(fatal bool, value bool, logs ...interface{}) {
	if !value {
		ast.logCaller()
		if len(logs) > 0 {
			ast.Log(logs...)
		} else {
			ast.Logf("value is not true")
		}
		ast.failIt(fatal)
	}
}

func (ast *Assert) Equal(expected, actual interface{}, logs ...interface{}) {
	ast.equalSprintAssert(true, true, expected, actual, logs...)
}

func (ast *Assert) NotEqual(expected, actual interface{}, logs ...interface{}) {
	ast.equalSprintAssert(true, false, expected, actual, logs...)
}

func (ast *Assert) equalSprintAssert(fatal bool, isEqual bool, expected, actual interface{}, logs ...interface{}) {
	expectedStr := fmt.Sprint(expected)
	actualStr := fmt.Sprint(actual)
	if isEqual != (expectedStr == actualStr) {
		ast.logCaller()
		if len(logs) > 0 {
			ast.Log(logs...)
		} else {
			if isEqual {
				ast.Log("Values not equal")
			} else {
				ast.Log("Values equal")
			}
		}
		ast.Log("Expected: ", expected)
		ast.Log("Actual: ", actual)
		ast.failIt(fatal)
	}
}

func (ast *Assert) logCaller() {
	_, file, line, _ := runtime.Caller(3)
	ast.Logf("Caller: %v:%d", file, line)
}

func (ast *Assert) failIt(fatal bool) {
	if fatal {
		ast.FailNow()
	} else {
		ast.Fail()
	}
}
