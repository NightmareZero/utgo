package utilp

// 伪三元表达式
// PS: 请注意，由于go的机制, trueVal, falseVal 都是已定值(在传入if前就确定了), 可能会导致panic
func If[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

// 如果data为nil，则返回0值，否则返回data
func OrZero[T any](data *T) T {
	if data == nil {
		var zero T
		switch any(zero).(type) {
		case int:
			return any(0).(T)
		case int8:
			return any(int8(0)).(T)
		case int16:
			return any(int16(0)).(T)
		case int32:
			return any(int32(0)).(T)
		case int64:
			return any(int64(0)).(T)
		case uint:
			return any(uint(0)).(T)
		case uint8:
			return any(uint8(0)).(T)
		case uint16:
			return any(uint16(0)).(T)
		case uint32:
			return any(uint32(0)).(T)
		case uint64:
			return any(uint64(0)).(T)
		case float32:
			return any(float32(0)).(T)
		case float64:
			return any(0.0).(T)
		case string:
			return any("").(T)
		case bool:
			return any(false).(T)
		default:
			return zero
		}
	}
	return *data
}
