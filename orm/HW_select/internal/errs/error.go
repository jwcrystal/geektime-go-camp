package errs

import (
	"errors"
	"fmt"
)

var (
	ErrPointerOnly = errors.New("orm: 只支持指向結構體的一級指針")
	ErrNoRows      = errors.New("orm: no data")
)

func NewErrUnknownField(name string) error {
	return fmt.Errorf("orm: unknown field %s", name)
}

// @ErrUnsupportedExpression 40001 原因是你输入了乱七八糟的类型
// 解决方案：使用正确的类型
func NewErrUnsupportedExpression(expr any) error {
	return fmt.Errorf("orm: 不支持表達式類型 %v", expr)
}

func NewErrUnsupportedSelectable(exp any) error {
	return fmt.Errorf("orm: 不支持目標列 %v", exp)
}
