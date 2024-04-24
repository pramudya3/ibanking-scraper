package tool

import (
	"reflect"
	"runtime"
)

func GetFuncInfo(v interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(v).Pointer()).Name()
}
