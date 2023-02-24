package util

// 伪三元表达式
// PS: 请注意，由于go的机制, trueVal, falseVal 都是已定值(在传入if前就确定了), 可能会导致panic
func If[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}
