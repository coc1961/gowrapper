package example

import (
	"github.com/coc1961/gowrapper/example/t1"
)

type IFace1 interface {
	IFace
	Func6(s string) error
	Func7(s string)
}

type IFace interface {
	Func5(s string) error
}

type TestInterface interface {
	IFace1
	Func4(string) error
	Func3(str1, str2 string) error
	Func2(arr []int) (IFace, t1.T1Interface, map[string]string, error)
	Func1(i int, pt1 t1.T1) (rt1 *t1.T1, xx t1.T1Interface, err error)
}
