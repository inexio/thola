package utility

// IfThenElse is a wrapper for the if condition.
func IfThenElse(condition bool, t interface{}, e interface{}) interface{} {
	if condition {
		return t
	}
	return e
}

// IfThenElseInt is a wrapper for the if condition.
func IfThenElseInt(condition bool, t int, e int) int {
	if condition {
		return t
	}
	return e
}

// IfThenElseString is a wrapper for the if condition.
func IfThenElseString(condition bool, t string, e string) string {
	if condition {
		return t
	}
	return e
}
